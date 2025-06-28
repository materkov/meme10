const { useState, useEffect } = React;

function App() {
  const [posts, setPosts] = useState([]);
  const [text, setText] = useState("");

  const fetchFeed = async () => {
    const resp = await fetch('/twirp/PostService/GetFeed', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}'
    });
    const data = await resp.json();
    if (data.posts) setPosts(data.posts);
  };

  useEffect(() => { fetchFeed(); }, []);

  const addPost = async () => {
    await fetch('/twirp/PostService/AddPost', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text })
    });
    setText('');
    fetchFeed();
  };

  return (
    <div>
      <div>
        <input value={text} onChange={e => setText(e.target.value)} />
        <button onClick={addPost}>Post</button>
      </div>
      <ul>
        {posts.map(p => (
          <li key={p.id} style={{textDecoration: p.deleted ? 'line-through' : 'none'}}>
            [{new Date(p.created_at).toLocaleString()}] {p.text}
          </li>
        ))}
      </ul>
    </div>
  );
}

ReactDOM.render(<App />, document.getElementById('root'));
