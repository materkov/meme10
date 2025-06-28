#!/usr/bin/env python3
import http.server
import socketserver
import os
from urllib.parse import urlparse

class SPAServer(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        # Parse the URL
        parsed_url = urlparse(self.path)
        path = parsed_url.path
        
        # If it's a file that exists, serve it normally
        if os.path.exists(path.lstrip('/')):
            return http.server.SimpleHTTPRequestHandler.do_GET(self)
        
        # Otherwise, serve index.html for SPA routing
        self.path = '/index.html'
        return http.server.SimpleHTTPRequestHandler.do_GET(self)

if __name__ == "__main__":
    PORT = 8000
    
    with socketserver.TCPServer(("", PORT), SPAServer) as httpd:
        print(f"SPA server running on http://localhost:{PORT}")
        print("All routes will fallback to index.html")
        httpd.serve_forever() 