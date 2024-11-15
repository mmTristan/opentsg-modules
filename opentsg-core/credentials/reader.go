// Package credentials returns the bytes from a http location
/*


func Example() {
	// Example Function
	// Generate a decoder before running the decode function.
	// All keys are optional.
	decoder, _ := AuthInit("s3profilename", "key1", "key2", "key3")
	// Then use the decoder to access the website with any of the added keys
	fileBytes, _ := decoder.Decode("example.com/pathto/important/file.json")
	//Do something with the extracted bytes
	fmt.Println(string(fileBytes))
}
*/
package credentials

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var regGitAPI = regexp.MustCompile(`https://gitlab\.com/api/v4/`)
var regGitL = regexp.MustCompilePOSIX(`https://gitlab\.com/`)

var regGitHbAPI = regexp.MustCompile(`https://api\.github\.com/`)
var regGitH = regexp.MustCompilePOSIX(`https://github\.com/`)

var regS3 = regexp.MustCompile(`^s3://[\w\-\.]{3,63}/`)
var regS3AWS = regexp.MustCompile(`^http://s3\.amazonaws\.com/[\w\-\.]{3,63}/`)

type contKey string

const credentialsAuth contKey = "test"

// InitDecoder adds the decoder body to context.
// it is to be used in tandem with the GetWebBytes function.
func InitDecoder(c *context.Context, s3Profile string, httpKeys ...string) error {

	authDecoder, err := AuthInit(s3Profile, httpKeys...)

	if err != nil {
		return err
	}

	*c = context.WithValue(*c, credentialsAuth, authDecoder)

	return nil
}

// GetWebBytes is a wrapper of `credentials` where the configuration body is stored in config.
// This is to prevent several intialisations of the authbody or the data being passed around.
func GetWebBytes(c *context.Context, uri string) ([]byte, error) {

	var ok = false
	var auth Decoder

	if c != nil {
		auth, ok = (*c).Value(credentialsAuth).(Decoder)
	}
	// if there is not an authorisation body make a new one with no credentials
	if !ok {
		var err error
		auth, err = AuthInit("")
		if err != nil {
			return nil, err
		}
	}
	return auth.Decode(uri)
}

// Decode returns the body of a url and an error if the information could not be extracted.
func (d Decoder) Decode(url string) ([]byte, error) {
	d.mut.Lock()
	defer d.mut.Unlock()
	// Insert a credentials manager
	tokenGen := d.authorisation
	switch {
	case regGitL.MatchString(url), regGitAPI.MatchString(url):
		return gitDecode(url, tokenGen["git_auth"])
	case regGitH.MatchString(url), regGitHbAPI.MatchString(url):
		return gitHubDecode(url, tokenGen["github_auth"])
		// develop functions for each regex string
	case regS3.MatchString(url), regS3AWS.MatchString(url):
		return s3Decode(url, tokenGen["s3_profile"])
		// Develop functions for each regex string
	default: // Make this for any other http decode
		return httpDecode(url)
	}

}

func httpDecode(url string) ([]byte, error) {
	// Look at implementing these
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	} else if resperr := repsonseHelper(resp); resperr != nil {
		return nil, resperr
	}
	return io.ReadAll(resp.Body)
}

type access struct {
	Content string `json:"content,omitempty"`
	Name    string `json:"name,omitempty"`
}

type jsonID struct {
	ID      int    `json:"id,omitempty"`
	Default string `json:"default_branch,omitempty"`
}

func gitDecode(url string, a token) ([]byte, error) {

	// Get the body of the gitlab api
	token := "Bearer " + a.tokenCode.(string)

	// Convert gitlab links to gitlab api calls
	if regGitL.MatchString(url) && !regGitAPI.MatchString(url) {
		// Get the owner and the repo
		owner, wantrepo := bucketToString(url, 19)
		// Split the repo into repo and file
		repo, file := bucketToString(wantrepo, 0)

		// Extract the api json
		idGetURL := "https://gitlab.com/api/v4/projects/" + owner + "%2f" + repo
		idGetJSON, err := getRequest(idGetURL, token)
		if err != nil {
			return nil, err
		}

		var id jsonID
		err = json.Unmarshal(idGetJSON, &id)
		if err != nil {
			return nil, fmt.Errorf("error gettting json of %s: %v", url, err)
		}

		if id.ID == 0 {
			return nil, fmt.Errorf("error no valid id found for %v", url)
		}
		newfile := strings.ReplaceAll(file, "/", "%2F")
		url = "https://gitlab.com/api/v4/projects/" + fmt.Sprintf("%v", id.ID) + "/repository/files/" + newfile + "?ref=" + id.Default
	}

	body, err := getRequest(url, token)
	if err != nil {
		return nil, err
	}

	// Decode from a json with base 64 to the actual contents of the file
	var file access
	err = json.Unmarshal(body, &file)
	if err != nil {
		return nil, fmt.Errorf("error gettting json of %s: %v", url, err)
	}

	dst := make([]byte, base64.StdEncoding.DecodedLen(len(file.Content)))
	_, err = base64.StdEncoding.Decode(dst, []byte(file.Content))

	return dst, err
}

func getRequest(url, token string) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	// set a token if one is provided
	if token != "" {
		// Req.Header.Set("PRIVATE-TOKEN", token)
		req.Header.Set("Authorization", token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	} else if resperr := repsonseHelper(resp); resperr != nil {
		return nil, resperr
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func gitHubDecode(url string, a token) ([]byte, error) {
	var token string
	if a.tokenCode.(string) == "" {
		token = a.tokenCode.(string)
	} else {
		token = "token " + a.tokenCode.(string)
	}

	// If not api then we'll amend it to that
	// https://github.com/mrmxf/ascmhl/schema/ascmhl.xsd
	// https://api.github.com/repos/mmTristan/ascmhl/contents/schema%2Fascmhl.xsd
	// Convert gitlab links to gitlab api calls
	if regGitH.MatchString(url) && !regGitHbAPI.MatchString(url) {
		// get the owner and the repo
		owner, wantrepo := bucketToString(url, 19)
		// split the repo into repo and file
		repo, file := bucketToString(wantrepo, 0)

		newfile := strings.ReplaceAll(file, "/", "%2F")
		url = "https://api.github.com/repos/" + owner + "/" + repo + "/contents/" + newfile
	}

	body, err := getRequest(url, token)
	if err != nil {
		return nil, err
	}

	var file access
	err = json.Unmarshal(body, &file)
	if err != nil {
		return nil, err
	}

	dst := make([]byte, base64.StdEncoding.DecodedLen(len(file.Content)))
	_, err = base64.StdEncoding.Decode(dst, []byte(file.Content))

	return dst, err
}

func s3Decode(url string, a token) ([]byte, error) {
	opt := (a.tokenCode).(*s3AuthDetail)

	// https://s3.console.aws.amazon.com/s3/object/mmh-cache?region=eu-west-2&prefix=bot-tlh/dev/schema/addimageschema.json
	// http://s3.amazonaws.com/[bucket_name]/object/mhl.jdon
	// Split the string to the bucket and file for use with s3 sdk
	var bucket, file string
	if regS3.MatchString(url) {
		bucket, file = bucketToString(url, 5)
	} else {
		bucket, file = bucketToString(url, 24)
	}
	region := opt.region
	if region == "" {
		region = "eu-west-2"
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: opt.credential,
	})
	if err != nil {
		return nil, err
	}

	downloader := s3manager.NewDownloader(sess)

	// Download the item from the bucket and check for errors before returning
	buf := aws.NewWriteAtBuffer([]byte{})
	_, err = downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(file),
		})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// BucketToString splits the s3 url into a bucket and file for use with the aws sdk
// split up the other string type as well
func bucketToString(url string, start int) (string, string) {
	var bucket int
	if start > len(url) {
		// If the start is out of range then nip it in the bud
		return url, ""
	}
	for i, let := range url[start:] {
		// Search for the end of the bucket name
		if let == rune('/') {
			bucket = i + start

			break
		}
	}

	// There's some error checking to be done here
	return url[start:bucket], url[bucket+1:]
}

func repsonseHelper(resp *http.Response) error {
	stat := resp.Status
	valid := regexp.MustCompile("OK")
	if !valid.MatchString(stat) {
		return fmt.Errorf(stat)
	}

	return nil
}
