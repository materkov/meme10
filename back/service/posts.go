package service

import (
	"context"
	"sort"
	"sync"
	"time"

	pb "socialnet/proto/posts"

	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PostsServer struct {
	posts []*pb.Post
	mu    sync.Mutex
	next  int64
}

func NewPostsServer() *PostsServer {
	return &PostsServer{
		posts: []*pb.Post{},
		next:  1,
	}
}

// AddPost implements the Twirp PostService method for adding a post.
func (s *PostsServer) AddPost(ctx context.Context, req *pb.AddPostRequest) (*pb.AddPostResponse, error) {
	//author := ctx.Value("user_id").(string) // Assuming user ID is injected in context by middleware
	s.mu.Lock()
	defer s.mu.Unlock()
	post := &pb.Post{
		Id:        s.next,
		AuthorId:  "1",
		Text:      req.Text,
		CreatedAt: timestamppb.New(time.Now()),
	}
	s.next++
	s.posts = append(s.posts, post)

	return &pb.AddPostResponse{Post: post}, nil
}

// DeletePost implements the Twirp PostService method for deleting a post.
func (s *PostsServer) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, post := range s.posts {
		if post.Id == req.Id {
			post.Deleted = true
			return &pb.DeletePostResponse{Post: post}, nil
		}
	}
	return nil, twirp.NotFoundError("post not found")
}

// GetFeed implements the Twirp PostService method for getting the feed.
func (s *PostsServer) GetFeed(ctx context.Context, req *pb.GetFeedRequest) (*pb.GetFeedResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]*pb.Post, len(s.posts))
	copy(list, s.posts)
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.AsTime().After(list[j].CreatedAt.AsTime())
	})
	return &pb.GetFeedResponse{Posts: list}, nil
}
