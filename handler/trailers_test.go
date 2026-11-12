package handler

import "testing"

func TestMovieSourcesNativeAndAdaptive(t *testing.T) {
	full:=movieSources(steamMovie{ID:1,Name:"Gameplay",MP4:map[string]string{"max":"http://video.fastly.steamstatic.com/store_trailers/123/movie_max.mp4"}},730)
	if full.Preview || len(full.Sources)!=1 || full.Sources[0]!="https://video.fastly.steamstatic.com/store_trailers/123/movie_max.mp4" { t.Fatalf("full: %#v",full) }
	m:=steamMovie{ID:2,HLS:"https://video.fastly.steamstatic.com/store_trailers/730/123/aa5a28c78f12232e6b6839034550c28b162fad3e/1748810724/hls_264_master.m3u8"}
	short:=movieSources(m,730)
	if !short.Preview || len(short.Sources)!=2 || short.Sources[0]!="https://video.fastly.steamstatic.com/store_trailers/730/123/aa5a28c78f12232e6b6839034550c28b162fad3e/1748810724/microtrailer.mp4" { t.Fatalf("preview: %#v",short) }
	if len(movieSources(m,570).Sources)!=0 { t.Fatal("must not link a different game's preview") }
	for _,raw:=range []string{"https://evil.example/store_trailers/a.mp4","https://steamstatic.com.evil.example/store_trailers/a.mp4","https://user@video.fastly.steamstatic.com/store_trailers/a.mp4","https://video.fastly.steamstatic.com/elsewhere/a.mp4"} {
		if steamVideoURL(raw)!="" { t.Fatalf("accepted %s",raw) }
	}
}

func TestParseSteamMoviesEmptyAndMalformed(t *testing.T) {
	for _, body := range []string{
		`{"70":{"success":true,"data":{"name":"Half-Life"}}}`,
		`{"70":{"success":true,"data":{"movies":[]}}}`,
	} {
		movies, err := parseSteamMovies([]byte(body), "70")
		if err != nil || len(movies) != 0 { t.Fatalf("movies=%#v err=%v", movies, err) }
	}
	for _, body := range []string{
		`{"70":{"success":true,"data":[]}}`,
		`{"70":{"success":false}}`,
		`{"70":{"success":true,"data":null}}`,
		`{"71":{"success":true,"data":{}}}`,
		`<html>error</html>`,
	} {
		if _,err := parseSteamMovies([]byte(body), "70"); err == nil { t.Fatalf("accepted %s", body) }
	}
}
