const { useState, useEffect } = React;

function App() {
  const [posts, setPosts] = useState([]);
  const [text, setText] = useState("");

  const handleVKCallback = async () => {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    
    if (code) {
      try {
        const response = await fetch('http://localhost:8080/twirp/socialnet.AuthService/VKAuth', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ code })
        });
        const authResponse = await response.json();
        console.log('VK Auth successful:', authResponse);
        // Clear the URL parameters after successful auth
        window.history.replaceState({}, document.title, window.location.pathname);
      } catch (error) {
        console.error('VK Auth failed:', error);
      }
    }
  };

  const getAuthUrl = async () => {
    const response = await fetch('http://localhost:8080/twirp/socialnet.AuthService/GetVKAuthURL', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}'
    });
    const urlResponse = await response.json();
    window.location.href = urlResponse.url;
  };

  const fetchFeed = async () => {
    const resp = await fetch('http://localhost:8080/twirp/socialnet.PostService/GetFeed', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}'
    });
    const data = await resp.json();
    if (data.posts) setPosts(data.posts);
  };

  useEffect(() => { 
    fetchFeed(); 
    handleVKCallback();
  }, []);

  const addPost = async () => {
    await fetch('http://localhost:8080/twirp/socialnet.PostService/AddPost', {
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
        <button onClick={getAuthUrl}>Authorize</button>
        <br/>
        <br/>
        <br/>

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
