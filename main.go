package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/djherbis/times"
)

var dir = flag.String("dir", ".", "The directory to read.")

type Track struct {
	MP3   string
	Title string
	Desc  string
	Made  string
	Tags  []string
}

type Section struct {
	Dir    string
	Title  string
	Desc   string
}

type TOMLTrack struct {
	Title   string
	Desc    string
	Section string
}

type TOMLSections struct {
	Section []*struct {
		Dir   string
		Title string
		Desc  string
	}
}

func main() {
	flag.Parse()

	ts := TOMLSections{}
	_, err := toml.DecodeFile(filepath.Join(*dir, "sections.toml"), &ts)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	var size int64 = 0

	tracks := []Track{}
	sections := []*Section{}
	for _, s := range ts.Section {
		if s.Title == "" {
			log.Fatal("Section with no title")
		}
		section := &Section{
			Dir:    s.Dir,
			Title:  s.Title,
			Desc:   s.Desc,
		}
		sections = append(sections, section)

		mp3files, err := filepath.Glob(filepath.Join(*dir, section.Dir, "*.mp3"))
		if err != nil {
			log.Fatal(err)
		}

		for _, mp3file := range mp3files {
			log.Println("Looking at mp3:", mp3file)
			base := filepath.Base(mp3file)
			clean := base[0 : len(base)-4] //remove .mp3
			meta := filepath.Join(*dir, section.Dir, clean+".toml")

			ft, err := times.Stat(mp3file)
			if err != nil {
				log.Fatal(err)
			}

			tt := TOMLTrack{}
			if _, err := os.Stat(meta); err == nil {
				log.Println("Found meta")
				_, err := toml.DecodeFile(meta, &tt)
				if err != nil {
					log.Fatal(err)
				}
			}
			stat, err := os.Stat(mp3file)
			if err != nil {
				log.Fatal(err)
			}
			size += stat.Size()

			var tags []string
			if s.Dir == "drone" {
				tags = []string{"noise", "dark"}
			} else {
				tags = []string{s.Dir}
			}

			track := Track{
				MP3:   base,
				Title: tt.Title,
				Desc:  tt.Desc,
				Made:  ft.ModTime().Format(time.RFC3339),
				Tags:  tags,
			}
			if tt.Title == "" {
				track.Title = clean
			}
			log.Printf("Adding %s to %s", track.Title, section.Title)
			tracks = append(tracks, track)
			count += 1
		}
	}

	out, err := os.Create("tracks.json")
	if err != nil {
		log.Fatalf("error opening output file: %s", err)
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	err = enc.Encode(tracks)
	if err != nil {
		log.Fatalf("error writing output file: %s", err)
	}


	log.Printf("%d tracks totaling %d MB", count, size/1024/1024)
}
