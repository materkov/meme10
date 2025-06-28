package service

import (
	"context"
	"testing"
	"time"

	pb "socialnet/proto/posts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostsServer(t *testing.T) {
	server := NewPostsServer()

	assert.NotNil(t, server)
	assert.Equal(t, int64(1), server.next)
	assert.Empty(t, server.posts)
}

func TestPostsServer_AddPost(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	// Test adding first post
	req := &pb.AddPostRequest{Text: "Hello, world!"}
	resp, err := server.AddPost(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Post)
	assert.Equal(t, int64(1), resp.Post.Id)
	assert.Equal(t, "1", resp.Post.AuthorId)
	assert.Equal(t, "Hello, world!", resp.Post.Text)
	assert.False(t, resp.Post.Deleted)
	assert.WithinDuration(t, time.Now(), resp.Post.CreatedAt.AsTime(), 2*time.Second)

	// Test adding second post
	req2 := &pb.AddPostRequest{Text: "Second post"}
	resp2, err := server.AddPost(ctx, req2)

	require.NoError(t, err)
	assert.Equal(t, int64(2), resp2.Post.Id)
	assert.Equal(t, "Second post", resp2.Post.Text)

	// Verify posts are stored
	assert.Len(t, server.posts, 2)
}

func TestPostsServer_AddPost_EmptyText(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	req := &pb.AddPostRequest{Text: ""}
	resp, err := server.AddPost(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "", resp.Post.Text)
}

func TestPostsServer_DeletePost(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	// Add a post first
	req := &pb.AddPostRequest{Text: "Test post"}
	resp, _ := server.AddPost(ctx, req)
	postID := resp.Post.Id

	// Test deleting the post
	deleteReq := &pb.DeletePostRequest{Id: postID}
	deleteResp, err := server.DeletePost(ctx, deleteReq)

	require.NoError(t, err)
	assert.NotNil(t, deleteResp)
	assert.True(t, deleteResp.Post.Deleted)
	assert.Equal(t, postID, deleteResp.Post.Id)

	// Verify the post is marked as deleted in storage
	assert.True(t, server.posts[0].Deleted)
}

func TestPostsServer_DeletePost_NotFound(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	// Try to delete a non-existent post
	req := &pb.DeletePostRequest{Id: 999}
	resp, err := server.DeletePost(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	// Check if it's a twirp NotFoundError
	assert.Contains(t, err.Error(), "post not found")
}

func TestPostsServer_DeletePost_EmptyServer(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	req := &pb.DeletePostRequest{Id: 1}
	resp, err := server.DeletePost(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPostsServer_GetFeed(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	// Add multiple posts
	server.AddPost(ctx, &pb.AddPostRequest{Text: "First post"})
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	server.AddPost(ctx, &pb.AddPostRequest{Text: "Second post"})
	time.Sleep(10 * time.Millisecond)
	server.AddPost(ctx, &pb.AddPostRequest{Text: "Third post"})

	// Get feed
	req := &pb.GetFeedRequest{}
	resp, err := server.GetFeed(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Posts, 3)

	// Verify posts are sorted by creation time (newest first)
	assert.Equal(t, "Third post", resp.Posts[0].Text)
	assert.Equal(t, "Second post", resp.Posts[1].Text)
	assert.Equal(t, "First post", resp.Posts[2].Text)

	// Verify IDs are correct
	assert.Equal(t, int64(3), resp.Posts[0].Id)
	assert.Equal(t, int64(2), resp.Posts[1].Id)
	assert.Equal(t, int64(1), resp.Posts[2].Id)
}

func TestPostsServer_GetFeed_Empty(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	req := &pb.GetFeedRequest{}
	resp, err := server.GetFeed(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Posts)
}

func TestPostsServer_GetFeed_WithDeletedPosts(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	// Add posts
	server.AddPost(ctx, &pb.AddPostRequest{Text: "First post"})
	server.AddPost(ctx, &pb.AddPostRequest{Text: "Second post"})

	// Delete first post
	server.DeletePost(ctx, &pb.DeletePostRequest{Id: 1})

	// Get feed
	req := &pb.GetFeedRequest{}
	resp, err := server.GetFeed(ctx, req)

	require.NoError(t, err)
	assert.Len(t, resp.Posts, 2)

	// Check deleted status by ID
	for _, post := range resp.Posts {
		if post.Id == 1 {
			assert.True(t, post.Deleted)
		} else if post.Id == 2 {
			assert.False(t, post.Deleted)
		}
	}
}

func TestPostsServer_ConcurrentAccess(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	// Test concurrent post creation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			req := &pb.AddPostRequest{Text: "Concurrent post"}
			_, err := server.AddPost(ctx, req)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all posts were created
	assert.Len(t, server.posts, 10)

	// Verify IDs are unique and sequential
	ids := make(map[int64]bool)
	for _, post := range server.posts {
		assert.False(t, ids[post.Id], "Duplicate ID found: %d", post.Id)
		ids[post.Id] = true
	}
}

func TestPostsServer_ConcurrentReadWrite(t *testing.T) {
	server := NewPostsServer()
	ctx := context.Background()

	// Add initial post
	server.AddPost(ctx, &pb.AddPostRequest{Text: "Initial post"})

	// Test concurrent read and write operations
	done := make(chan bool, 5)

	// Start read operations
	for i := 0; i < 3; i++ {
		go func() {
			req := &pb.GetFeedRequest{}
			_, err := server.GetFeed(ctx, req)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Start write operations
	for i := 0; i < 2; i++ {
		go func() {
			req := &pb.AddPostRequest{Text: "Concurrent post"}
			_, err := server.AddPost(ctx, req)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all operations to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify final state
	assert.Len(t, server.posts, 3) // 1 initial + 2 concurrent
}
