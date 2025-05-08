package credentials

// Use the convey method of testing
import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/url"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
)

func init() {
	s3Access = os.Getenv("AWS_ACCESS_KEY_ID")
	s3Region = os.Getenv("AWS_DEFAULT_REGION")
	s3Secret = os.Getenv("AWS_SECRET_ACCESS_KEY")
	ghToken = os.Getenv("GITHUB_PAT")
	glToken = os.Getenv("GITLAB_PAT")
}

var ghToken, glToken, s3Access, s3Secret, s3Region string

func TestDecodeHTTP(t *testing.T) {

	// Test getting http images with known urls
	addresses := []string{"https://opentsg.studio/blog/2023/09/13/2023-09-13-coming-soon/featured-logo-otsg.png",
		"https://mrmxf.io/smpte-ra-demo/register/lmt/schema"}
	expec := []string{"a673f3a297c99e270f8dabb1478e86516187ece596a3b0afe378db7bb79dbdb5",
		"f0537bd22e23f61dac9c1abcd8cdfa8dcf1d7b907bedb62fe55ae8b653ec525c"}

	// Open the url and check it is good
	emptyDecode, _ := AuthInit("")

	for i, ad := range addresses {
		htest := sha256.New()
		ftest, err := emptyDecode.Decode(ad)

		htest.Write(ftest)
		// Generate a sha of the file
		// fmt.Println(string(ftest))
		Convey("Checking that json and image files are extracted with http", t, func() {
			Convey(fmt.Sprintf("using a website of %v", ad), func() {
				Convey("A matching hash of a file is returned", func() {
					So(err, ShouldBeNil)
					So(fmt.Sprintf("%x", htest.Sum(nil)), ShouldResemble, expec[i])
				})
			})
		})
	}
}

func TestDecodeGitHub(t *testing.T) {
	// tokenB, err := os.ReadFile("./testdata/ghkey.txt")
	token := string(ghToken)
	// Test getting http images with known urls
	if ghToken != "" {
		addresses := []string{"https://api.github.com/repos/mmTristan/public/contents/nested%2Fnest.json",
			"https://api.github.com/repos/mmTristan/ascmhl/contents/schema%2Fascmhl.xsd",
			"https://github.com/mrmxf/ascmhl/schema/ascmhl.xsd"}
		tokens := []string{"", token, token}
		expec := []string{"bb9dd8180d70abce882ccdb69aab2bffa1a96c8f86cdbcd631948b3085465ab9",
			"0235e307f3930b9f8142c37f23f6e4d55fedb534dda19227069c2c95948f2bcb",
			"0235e307f3930b9f8142c37f23f6e4d55fedb534dda19227069c2c95948f2bcb"}

		// Open the url and check it is good

		for i, ad := range addresses {
			htest := sha256.New()
			genDec, _ := AuthInit("", tokens[i])
			ftest, err := genDec.Decode(ad)
			// fmt.Println(string(ftest))
			htest.Write(ftest)
			// Generate a sha of the file
			// fmt.Println(string(ftest))
			Convey("Checking that json and image files are extracted with github access", t, func() {
				Convey(fmt.Sprintf("using a website of %v", ad), func() {
					Convey("A matching hash of the extracted file is returned", func() {
						So(err, ShouldBeNil)
						So(fmt.Sprintf("%x", htest.Sum(nil)), ShouldResemble, expec[i])
					})
				})
			})
		}
	} else {
		// fail the test here so the environment should be set up
		fmt.Printf("github tests skipped due to the following errors %v opening the token file\n", "oh no")
	}
}

func TestDecodeGit(t *testing.T) {
	// Access the token so it is not saved in the test suite

	token := string(glToken)
	if token != "" {
		// Test getting http images with known urls
		addresses := []string{"https://gitlab.com/api/v4/projects/35946043/repository/files/go.mod?ref=main",
			"https://gitlab.com/api/v4/projects/33185381/repository/files/test%2Ftestapi.json?ref=main",
			"https://gitlab.com/mmTristan/publicgo/go.mod",
			"https://gitlab.com/mmTristan/basic-12bit-clour/test/testapi.json"}
		tokens := []string{"", token, "", token}
		expec := []string{"07f2ba17f973dda06607381921ab8811c177b0d2910e2d252e02cc0946b4cc7d",
			"87bad426555c30caa55b4332304f55e1ae79d331297ba75ab244539018798fbd",
			"07f2ba17f973dda06607381921ab8811c177b0d2910e2d252e02cc0946b4cc7d",
			"87bad426555c30caa55b4332304f55e1ae79d331297ba75ab244539018798fbd"}

		// Open the url and check it is good

		for i, ad := range addresses {
			htest := sha256.New()
			genDec, _ := AuthInit("", tokens[i])
			ftest, err := genDec.Decode(ad)

			htest.Write(ftest)

			Convey("Checking that json and image files are extracted from gitlab addresses", t, func() {
				Convey(fmt.Sprintf("using a website the gitlab api of %v", ad), func() {
					Convey("A matching hash of the extracted file is returned", func() {
						So(err, ShouldBeNil)
						So(fmt.Sprintf("%x", htest.Sum(nil)), ShouldResemble, expec[i])
					})
				})
			})
		}
	} else {
		fmt.Printf("gitlab tests skipped due to the following errors %v opening the token file\n", "oh no")
	}
}

func TestDecodeS3(t *testing.T) {
	// Extract the access details for the test
	region := s3Region
	secret := string(s3Secret)
	access := string(s3Access)
	// Run the test if there are no errors getting the details
	if secret != "" && access != "" {
		// Access the token so it is not saved in the test suite
		addresses := []string{"s3://mmh-cache/bot-tlh/staging/publish/multiramp-12b-pc-4k-zp.dpx", "http://s3.amazonaws.com/mmh-cache/bot-tlh/dev/schema/addimageschema.json"}
		expec := []string{"7490fd92c6292b6850bd6fd568abbe8bebae1a4ef3382c284d252921d94a6d4d",
			"679505dad56fb9aad50089ebb9e5b217912c41a9b585217f467633749bcbddac"}

		for i, ad := range addresses {
			htest := sha256.New()
			genDec, _ := AuthInit("", region, access, secret)
			ftest, err := genDec.Decode(ad)
			htest.Write(ftest)
			Convey("Checking that json and image files are extracted from s3 buckets", t, func() {
				Convey(fmt.Sprintf("using an s3 link of %v", ad), func() {
					Convey("A matching hash of the extracted file is returned", func() {
						So(err, ShouldBeNil)
						So(fmt.Sprintf("%x", htest.Sum(nil)), ShouldResemble, expec[i])
					})
				})
			})
		}
		address := "s3://mmh-cache/bot-tlh/staging/publish/multiramp-12b-pc-4k-zp.dpx"
		expecHash := "7490fd92c6292b6850bd6fd568abbe8bebae1a4ef3382c284d252921d94a6d4d"

		passed := func() (string, string, string) { return region, secret, access }
		envGet := func() (string, string, string) {

			return "", "", ""
		}
		cred := func() (string, string, string) {
			// Generate all of the files necessary
			f, _ := os.Create(config.DefaultSharedCredentialsFilename()) // , os.O_RDWR|os.O_CREATE, 0755)
			// fmt.Println(err)
			defer f.Close()
			file := "[default]\n" + "aws_access_key_id=" + access + "\n" +
				"aws_secret_access_key=" + secret
			_, _ = f.Write([]byte(file))
			// fmt.Println(err, "DEFAULT")

			return "", "", ""
		}

		snapShot := afero.NewOsFs()
		snapKey := os.Getenv("AWS_ACCESS_KEY_ID")
		snapReg := os.Getenv("AWS_DEFAULT_REGION")
		snapSecret := os.Getenv("AWS_SECRET_ACCESS_KEY")
		// After the test has run, reset all values back to the snap shot
		accessorMethods := []func() (string, string, string){passed, cred, envGet} // cred, envGet}
		accessorString := []string{"manually passed keys", "A credentials file", "extracting the environment variables"}

		for i, access := range accessorMethods {
			htest := sha256.New()
			// reset the env each time
			os.Setenv("AWS_ACCESS_KEY_ID", "")
			os.Setenv("AWS_DEFAULT_REGION", "")
			os.Setenv("AWS_SECRET_ACCESS_KEY", "")

			passed1, passed2, passed3 := access()

			genDec, _ := AuthInit("", passed1, passed2, passed3)
			ftest, err := genDec.Decode(address)
			htest.Write(ftest)
			Convey("Checking that different authentication methods for s3 work", t, func() {
				Convey(fmt.Sprintf("using an s3 link of %v and the method of %s", address, accessorString[i]), func() {
					Convey("A matching hash of the extracted file is returned", func() {
						So(err, ShouldBeNil)
						So(fmt.Sprintf("%x", htest.Sum(nil)), ShouldResemble, expecHash)
					})
				})
			})
		}
		os.Setenv("AWS_ACCESS_KEY_ID", snapKey)
		os.Setenv("AWS_DEFAULT_REGION", snapReg)
		os.Setenv("AWS_SECRET_ACCESS_KEY", snapSecret)
		f, _ := os.Create(config.DefaultSharedCredentialsFilename())
		defer f.Close()
		oldBody, _ := snapShot.Open(config.DefaultSharedCredentialsFilename())
		old, _ := io.ReadAll(oldBody)
		_, _ = f.Write(old)

		// Test authorisation errors make a array of functions that return strings

	} else {
		fmt.Printf("s3 tests skipped due to the following errors %v, %v opening the token files\n", " errS, errA, ", "on")
	}
}

func TestErrors(t *testing.T) {

	// Test getting http images with known urls
	baddresses := []string{"not even a website", "https://a.really.fake.website/not/real",
		"https://gitlab.com/api/v4/projects/33186381/repository/files/test%2Ftestapi.json?ref=main",
		"https://mrmxf.com/supersecret", "https://mrmxf.com/user"}

	want := []error{fmt.Errorf("Get \"not%%20even%%20a%%20website\": unsupported protocol scheme \"\""),
		fmt.Errorf(`Get "https://a.really.fake.website/not/real": dial tcp: lookup a.really.fake.website on 10.255.255.254:53: no such host`),
		fmt.Errorf("error received a status of 404 Not Found"), fmt.Errorf("404 Not Found"), fmt.Errorf("401 Unauthorized")}

	// Open the url and check it is good

	emptyDecode, _ := AuthInit("")

	for i, ad := range baddresses {
		// Open the files to be tested

		_, err := emptyDecode.Decode(ad)
		if uErr, ok := err.(*url.Error); ok {

			err = fmt.Errorf("%v", uErr.Error())
		}

		Convey("Checking that errors are returned for incorrect websites giving reasons for no access", t, func() {
			Convey(fmt.Sprintf("using a website of %v", ad), func() {
				Convey(fmt.Sprintf("An error of %v is returned", want[i]), func() {
					So(err, ShouldResemble, want[i])
				})
			})
		})
	}
}
