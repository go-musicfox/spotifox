package mercury

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/golang/protobuf/proto"
	//"google.golang.org/protobuf/proto"

	"github.com/arcspace/go-librespot/Spotify"
)

func (m *Client) mercuryGet(url string) ([]byte, error) {
	done := make(chan []byte)
	errs := make(chan error)
	go func(){
		err := m.Request(Request{
			Method:  "GET",
			Uri:     url,
			Payload: [][]byte{},
		}, func(res Response) {
			done <- res.CombinePayload()
		})
		if err != nil {
			errs <- err
		}
	}()
	select {
		case err := <-errs: return nil, err
		default:
	}
	result := <-done
	return result, nil
}

func (m *Client) mercuryGetJson(url string, result interface{}) error {
	data, err := m.mercuryGet(url)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

func (m *Client) mercuryGetProto(url string, result proto.Message) error {
	data, err := m.mercuryGet(url)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, result)
}

func (m *Client) GetRootPlaylist(username string) (*Spotify.SelectedListContent, error) {
	uri := fmt.Sprintf("hm://playlist/user/%s/rootlist", username)
	result := &Spotify.SelectedListContent{}
	err := m.mercuryGetProto(uri, result)
	return result, err
}

func (m *Client) GetPlaylist(id string) (*Spotify.SelectedListContent, error) {
	//uri := fmt.Sprintf("hm://playlist/%s", id)
	uri := fmt.Sprintf("hm://playlist/v2/playlist/%s", id)
	
	result := &Spotify.SelectedListContent{}
	err := m.mercuryGetProto(uri, result)
	return result, err
}

func (m *Client) GetToken(clientId string, scopes string) (*Token, error) {
	uri := fmt.Sprintf(
		"hm://keymaster/token/authenticated?client_id=%s&scope=%s",
		url.QueryEscape(clientId),
		url.QueryEscape(scopes),
	)
	token := &Token{}
	err := m.mercuryGetJson(uri, token)
	return token, err
}

func (m *Client) Search(search string, limit int, country string, username string) (*SearchResponse, error) {
	v := url.Values{}
	v.Set("entityVersion", "2")
	v.Set("limit", fmt.Sprintf("%d", limit))
	v.Set("imageSize", "large")
	v.Set("catalogue", "")
	v.Set("country", country)
	v.Set("platform", "zelda")
	v.Set("username", username)
	uri := fmt.Sprintf("hm://searchview/km/v4/search/%s?%s", url.QueryEscape(search), v.Encode())
	result := &SearchResponse{}
	err := m.mercuryGetJson(uri, result)
	return result, err
}

func (m *Client) Suggest(search string) (*SuggestResult, error) {
	uri := "hm://searchview/km/v3/suggest/" + url.QueryEscape(search) + "?limit=3&intent=2516516747764520149&sequence=0&catalogue=&country=&locale=&platform=zelda&username="
	data, err := m.mercuryGet(uri)
	if err != nil {
		return nil, err
	}
	return ParseSuggest(data)
}

func (m *Client) GetTrack(uri string) (trackID string, track *Spotify.Track, err error) {
	var hexID string
	trackID, hexID, err = Spotify.ExtractAssetID(uri)
	if err != nil {
		return
	}
	url := "hm://metadata/4/track/" + hexID
	track = &Spotify.Track{}
	err = m.mercuryGetProto(url, track)
	return
}

func (m *Client) GetArtist(uri string) (artistID string, artist*Spotify.Artist, err error) {
	var hexID string
	artistID, hexID, err = Spotify.ExtractAssetID(uri)
	if err != nil {
		return
	}
	url := "hm://metadata/4/artist/" + hexID
	artist = &Spotify.Artist{}
	err = m.mercuryGetProto(url, artist)
	return
}

func (m *Client) GetAlbum(uri string) (albumID string, album *Spotify.Album, err error) {
	var hexID string
	albumID, hexID, err = Spotify.ExtractAssetID(uri)
	if err != nil {
		return
	}
	
	url := "hm://metadata/4/album/" + hexID
	album =  &Spotify.Album{}
	err = m.mercuryGetProto(url, album)
	return
}

func ParseSuggest(body []byte) (*SuggestResult, error) {
	result := &SuggestResult{}
	err := json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	for _, s := range result.Sections {
		switch s.Typ {
		case "top-results":
			err = json.Unmarshal(s.RawItems, &result.TopHits)
		case "album-results":
			err = json.Unmarshal(s.RawItems, &result.Albums)
		case "artist-results":
			err = json.Unmarshal(s.RawItems, &result.Artists)
		case "track-results":
			err = json.Unmarshal(s.RawItems, &result.Tracks)
		}
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
