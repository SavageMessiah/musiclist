package main

import (
	"flag"
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/djherbis/times"
)

var dir = flag.String("dir", ".", "The directory to read.")

type Track struct {
	File    string
	Link    string
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
	Dir    string
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

	sections := []*Section{}
	for _, s := range ts.Section {
		if s.Title == "" {
			log.Fatal("Section with no title")
		}
		section := &Section{
			Dir:    s.Dir,
			Title:  s.Title,
			Desc:   s.Desc,
			Tracks: []Track{},
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

			track := Track{
				Title:   tt.Title,
				Desc:    tt.Desc,
				Link:    path.Join(section.Dir, base),
				File:    base,
				Made:    ft.ModTime().Format("Jan 2, 2006"),
				madeRaw: ft.ModTime(),
			}
			log.Printf("Adding %s to %s", track.Title, section.Title)
			section.Tracks = append(section.Tracks, track)
			count += 1
		}

		sort.Stable(section.Tracks)
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

	if err := t.Lookup("index.html.tmpl").Execute(out, sections); err != nil {
		log.Fatal(err)
	}

	log.Printf("%d tracks totaling %d MB", count, size/1024/1024)
}
