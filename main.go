package main

import (
	"net/http"
	"strconv"
	"html/template"
	"sync"
	"time"
	"math/rand"
	"encoding/json"
	"strings"
)

const (
	lag int = 42
)

type person struct {
	m	[2][lag][lag]bool
	k	int
	l	sync.Mutex
	n	bool
	id	string
	running bool
}

type persons struct {
	m map[string](*person)
}

var (
	r *rand.Rand
	mp persons
)

func getrand()bool {
	return r.Float32() > 0.7
}
///Persons Interface To Net : pprint, change
func (pss *persons)pprint(id string, w http.ResponseWriter, t string) {
	w.Write(([]byte)(t+"|"+pss.m[id].output()))
}

func (pps *persons)change(id string, info string, w http.ResponseWriter) {
	pps.m[id].iupdate(info)
}

func o(b bool)int {
	if b {
		return 1
	}
	return 0
}
///Person's Module : tupdate, iupdate, output
func (ps *person)tupdate(){
	ps.l.Lock()
	cc := 0
	ct := 0
	k := ps.k
	for i := 1; i < lag - 1; i++ {
		for j := 1; j < lag - 1; j++ {
			ct = 0
			ct = ct + o(ps.m[k][i - 1][j - 1]) + o(ps.m[k][i - 1][j]) + o(ps.m[k][i - 1][j + 1])
			ct = ct + o(ps.m[k][i][j - 1]) + o(ps.m[k][i][j + 1])
			ct = ct + o(ps.m[k][i + 1][j - 1]) + o(ps.m[k][i + 1][j]) + o(ps.m[k][i + 1][j + 1])
			cc += ct
			ps.m[1-k][i][j] = false
			if !ps.m[k][i][j] {
				if ct == 3  {
					ps.m[1-k][i][j] = true
				}
			}else{
				if ct == 2 || ct == 3 {
					ps.m[1-k][i][j] = true
				}
			}
		}
	}
	ps.k = 1 - k
	ps.l.Unlock()
}
func (ps *person)output()string {
	ps.l.Lock()
	ps.n = true
	var mtot [lag-2]int64
    for i := lag - 2; i > 0; i-- {
		k := mtot[i - 1]
		for j := lag - 2; j > 0; j-- {
			k = k * 2
			if ps.m[ps.k][i][j] {
				k++
			}
		}
		mtot[i - 1] = k
    }
	ps.l.Unlock()
	st, _ := json.Marshal(mtot)
	println(string(st))
	return string(st)
}

func getint(ss string)int {
	i, err := strconv.Atoi(ss)
	if err != nil {
		i = 0;
	}
	return i
}

func (ps *person)iupdate(xy string) {
	ps.l.Lock()
	ls := strings.Split(xy, "_")
	x := getint(ls[0])
	y := getint(ls[1])
	ps.m[ps.k][x][y] = true
	ps.n = true
	ps.l.Unlock()
}

func newPerson(id string)(*person) {
	d := &person {
		id: id,
		running: true,
	}

	for i := 1; i < lag - 1; i++ {
		for j := 1; j < lag - 1; j++ {
			d.m[0][i][j] = getrand()
		}
	}

	//Keep Modify 
	go func(p *person){
		for p.running {
			p.tupdate()
			time.Sleep(1e9)
		}
	}(d)
	//Judge Kill
	go func(p *person){
		for p.running{
			p.n = false
			time.Sleep(60e9)
			if !p.n {
				p.running = false
				delete(mp.m, p.id)
			}
		}
	}(d)
	return d
}
///Hands change auto lg
func changeHand(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if _, ok := r.Form["ch"]; !ok {
		w.Write([]byte{})
		return
	}
	ch := r.FormValue("ch")

	if _, ok := r.Form["id"]; !ok {
		w.Write([]byte{})
		return
	}
	id := r.FormValue("id")
	if _, ok := mp.m[id]; !ok {
		w.Write([]byte{})
		return
	}

	mp.change(id, ch, w)
}

func autoHand(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if _, ok := r.Form["id"]; !ok {
		w.Write([]byte{})
		return
	}
	id := r.FormValue("id")
	t, err := template.ParseFiles("html/auto.html")
	if err != nil {
		print(err.Error())
	}
	t.Execute(w, id)
}

func lgHand(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if _, ok := r.Form["id"]; !ok {
		w.Write([]byte{})
		return
	}
	id := r.FormValue("id")
	if _, ok := mp.m[id]; !ok {
		mp.m[id] = newPerson(id)
	}
	mp.pprint(id, w, r.FormValue("t"))
}

func main() {
	mp = persons {
		map[string](*person){},
	}
	r = rand.New(rand.NewSource(time.Now().Unix()))
	http.HandleFunc("/lg", lgHand)
	http.HandleFunc("/auto", autoHand)
	http.HandleFunc("/change", changeHand)
	http.Handle("/js/", http.FileServer(http.Dir("./html")))
	http.ListenAndServe(":8886", nil)
}
