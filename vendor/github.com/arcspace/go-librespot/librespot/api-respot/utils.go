package respot

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
)

const kAPEndpoint = "https://APResolve.spotify.com/"

// APList is the JSON structure corresponding to the output of the AP endpoint resolve API
type APList struct {
	ApList []string `json:"ap_list"`
}

// APResolve fetches the available Spotify access point URLs and picks a random one
func APResolve() (string, error) {
	r, err := http.Get(kAPEndpoint)
	if err != nil {
		return "", fmt.Errorf("failed to resolve Spotify access point lookup: %v", err)
	}
	defer r.Body.Close()


	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	var endpoints APList
	err = json.Unmarshal(body, &endpoints)
	if err != nil {
		return "", err
	}
	if len(endpoints.ApList) == 0 {
		return "", errors.New("AP endpoint list is empty")
	}

	return endpoints.ApList[rand.Intn(len(endpoints.ApList))], nil
}

func GenerateDeviceUID(name string) string {
	hash := sha1.Sum([]byte(name))
	hash64 := base64.StdEncoding.EncodeToString(hash[:])
	return hash64
}
