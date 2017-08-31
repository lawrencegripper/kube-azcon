package azureProviders

import (
	"time"
	"math/rand"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

func StringPointer(i string) *string { return &i }
func Int32Pointer(i int32) *int32    { return &i }
func Int64Pointer(i int64) *int64    { return &i }
func IntPointer(i int) *int          { return &i }

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
func NewServicePrincipalTokenFromCredentials(c ARMConfig, scope string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, c.TenantID)
	if err != nil {
		panic(err)
	}
	return adal.NewServicePrincipalToken(*oauthConfig, c.ClientID, c.ClientSecret, scope)
}

//This is probably a pretty nasty hack, was interesting to play with runes.
//Todo: Simplify Password generation.
var lettersLower = []rune("abcdefghijklmnopqrstuvwxyz")
var lettersUpper = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
var numbers = []rune("1234567890")
var symbols = []rune("!@£%$£^*_+")

func randAlphaNumericSeq(n int) string {
	bucketSize := n / 4
	return randFromSelection(bucketSize, lettersUpper) + randFromSelection(bucketSize, lettersLower) + randFromSelection(bucketSize, numbers) + randFromSelection(bucketSize, symbols)
}

func randFromSelection(length int, choices []rune) string {
	b := make([]rune, length)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = choices[rand.Intn(len(choices))]
	}
	return string(b)
}
