
Spotify
=======

[![GoDoc](https://godoc.org/github.com/zmb3/spotify?status.svg)](http://godoc.org/github.com/zmb3/spotify)

This is a Go wrapper for working with Spotify's
[Web API](https://developer.spotify.com/web-api/).

It aims to support every task listed in the Web API Endpoint Reference,
located [here](https://developer.spotify.com/web-api/endpoint-reference/).

By using this library you agree to Spotify's
[Developer Terms of Use](https://developer.spotify.com/developer-terms-of-use/).

## Installation

To install the library, simply

`go get github.com/zmb3/spotify/v2`

## Authentication

Spotify uses OAuth2 for authentication and authorization.  
As of May 29, 2017 _all_ Web API endpoints require an access token.

You can authenticate using a client credentials flow, but this does not provide
any authorization to access a user's private data.  For most use cases, you'll
want to use the authorization code flow.  This package includes an `Authenticator`
type to handle the details for you.

Start by registering your application at the following page:

https://developer.spotify.com/my-applications/.

You'll get a __client ID__ and __secret key__ for your application.  An easy way to
provide this data to your application is to set the SPOTIFY_ID and SPOTIFY_SECRET
environment variables.  If you choose not to use environment variables, you can
provide this data manually.


````Go
// the redirect URL must be an exact match of a URL you've registered for your application
// scopes determine which permissions the user is prompted to authorize
auth := spotifyauth.New(spotifyauth.WithRedirectURL(redirectURL), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate))

// get the user to this URL - how you do that is up to you
// you should specify a unique state string to identify the session
url := auth.AuthURL(state)

// the user will eventually be redirected back to your redirect URL
// typically you'll have a handler set up like the following:
func redirectHandler(w http.ResponseWriter, r *http.Request) {
      // use the same state string here that you used to generate the URL
      token, err := auth.Token(r.Context(), state, r)
      if err != nil {
            http.Error(w, "Couldn't get token", http.StatusNotFound)
            return
      }
      // create a client using the specified token
      client := spotify.New(auth.Client(r.Context(), token))

      // the client can now be used to make authenticated requests
}
````

You may find the following resources useful:

1. Spotify's Web API Authorization Guide:
https://developer.spotify.com/web-api/authorization-guide/

2. Go's OAuth2 package:
https://godoc.org/golang.org/x/oauth2/google


## Helpful Hints

### Automatic Retries

The API will throttle your requests if you are sending them too rapidly.
The client can be configured to wait and re-attempt the request.
To enable this, set the `AutoRetry` field on the `Client` struct to `true`.

For more information, see Spotify [rate-limits](https://developer.spotify.com/web-api/user-guide/#rate-limiting).

## API Examples

Examples of the API can be found in the [examples](examples) directory.

You may find tools such as [Spotify's Web API Console](https://developer.spotify.com/web-api/console/)
or [Rapid API](https://rapidapi.com/package/SpotifyPublicAPI/functions?utm_source=SpotifyGitHub&utm_medium=button&utm_content=Vendor_GitHub)
valuable for experimenting with the API.

### Missing data in responses

It's extremely common that when there is no market information available in your
request, that the Spotify API will simply return null for details about a track
or episode.

This typically occurs when you are just using an application's auth token, and
aren't impersonating a user via oauth. As when you are using a token associated
with a user, the user's market seems to be extracted from their profile and
used when producing the response.
