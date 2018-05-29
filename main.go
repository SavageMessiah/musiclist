package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/djherbis/times"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

var dir = flag.String("dir", ".", "The directory to read.")

type Track struct {
	File    string
	Title   string
	Desc    string
	Made    string
	madeRaw time.Time
}

type Tracks []Track

func (t Tracks) Len() int           { return len(t) }
func (t Tracks) Less(i, j int) bool { return t[i].madeRaw.After(t[j].madeRaw) }
func (t Tracks) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

type Section struct {
	Title  string
	Desc   string
	Tracks Tracks
}

type TOMLTrack struct {
	Title   string
	Desc    string
	Section string
}

type TOMLSections struct {
	Section []*struct {
		Short string
		Title string
		Desc  string
	}
}

func main() {
	flag.Parse()

	mp3files, err := filepath.Glob(filepath.Join(*dir, "*.mp3"))
	if err != nil {
		log.Fatal(err)
	}

	ts := TOMLSections{}
	_, err = toml.DecodeFile(filepath.Join(*dir, "sections.toml"), &ts)
	if err != nil {
		log.Fatal(err)
	}

	sections := make(map[string]*Section)
	for _, s := range ts.Section {
		if s.Title == "" {
			log.Fatal("Section with no title")
		}
		if s.Short == "" {
			s.Short = s.Title
		}
		sections[s.Short] = &Section{
			Title:  s.Title,
			Desc:   s.Desc,
			Tracks: []Track{},
		}
	}
	for _, mp3file := range mp3files {
		log.Println("Looking at mp3:", mp3file)
		base := filepath.Base(mp3file)
		clean := base[0 : len(base)-4] //remove .mp3
		meta := filepath.Join(*dir, clean+".toml")

		ft, err := times.Stat(mp3file)
		if err != nil {
			log.Fatal(err)
		}

		tt := TOMLTrack{
			Section: "Uncategorized",
		}
		if _, err := os.Stat(meta); err == nil {
			log.Println("Found meta")
			_, err := toml.DecodeFile(meta, &tt)
			if err != nil {
				log.Fatal(err)
			}
		}

		track := Track{
			Title:   tt.Title,
			Desc:    tt.Desc,
			File:    base,
			Made:    ft.BirthTime().Format("Jan 2, 2006"),
			madeRaw: ft.BirthTime(),
		}
		section, ok := sections[tt.Section]
		if !ok {
			log.Fatalf("Missing section: %s", tt.Section)
		}
		log.Printf("Adding %s to %s", track.Title, tt.Section)
		section.Tracks = append(section.Tracks, track)
	}

	sectionsList := []Section{}

	for _, s := range ts.Section {
		section := sections[s.Short]
		sort.Stable(section.Tracks)
		sectionsList = append(sectionsList, *section)
	}

	out, err := os.Create(filepath.Join(*dir, "index.html"))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	t := template.New("main")
	if _, err := t.ParseFiles(filepath.Join(*dir, "index.html.tmpl")); err != nil {
		log.Fatal(err)
	}

	if err := t.Lookup("index.html.tmpl").Execute(out, sectionsList); err != nil {
		log.Fatal(err)
	}
}
