package service

import (
	"encoding/json"
	"net/http"
	"sort"
	"sync"
	"time"

	"socialnet/twirp"
)

// Post represents a single post in the system
type Post struct {
	ID        int64     `json:"id"`
	AuthorID  string    `json:"author_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	Deleted   bool      `json:"deleted"`
}

type Server struct {
	mux   *http.ServeMux
	posts []*Post
	mu    sync.Mutex
	next  int64
}

func NewServer() *Server {
	s := &Server{mux: http.NewServeMux(), posts: []*Post{}, next: 1}
	s.registerHandlers()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerHandlers() {
	s.mux.HandleFunc("/twirp/PostService/AddPost", s.handleAddPost)
	s.mux.HandleFunc("/twirp/PostService/DeletePost", s.handleDeletePost)
	s.mux.HandleFunc("/twirp/PostService/GetFeed", s.handleGetFeed)
}

func (s *Server) handleAddPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		twirp.WriteError(w, &twirp.Error{Code: "bad_route", Msg: "POST required"})
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		twirp.WriteError(w, &twirp.Error{Code: "malformed", Msg: "cannot decode"})
		return
	}
	author := r.Header.Get("X-User")
	if author == "" {
		twirp.WriteError(w, &twirp.Error{Code: "unauthenticated", Msg: "missing user"})
		return
	}
	s.mu.Lock()
	p := &Post{ID: s.next, AuthorID: author, Text: req.Text, CreatedAt: time.Now()}
	s.next++
	s.posts = append(s.posts, p)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(p)
}

func (s *Server) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		twirp.WriteError(w, &twirp.Error{Code: "bad_route", Msg: "POST required"})
		return
	}
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		twirp.WriteError(w, &twirp.Error{Code: "malformed", Msg: "cannot decode"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.posts {
		if p.ID == req.ID {
			p.Deleted = true
			_ = json.NewEncoder(w).Encode(p)
			return
		}
	}
	twirp.WriteError(w, &twirp.Error{Code: "not_found", Msg: "post not found"})
}

func (s *Server) handleGetFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		twirp.WriteError(w, &twirp.Error{Code: "bad_route", Msg: "POST required"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]*Post, len(s.posts))
	copy(list, s.posts)
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Posts []*Post `json:"posts"`
	}{list})
}
