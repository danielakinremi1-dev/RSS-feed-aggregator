# Renders docs/demo.gif with VHS (https://github.com/charmbracelet/vhs).
# Run it with `make demo`: that starts a throwaway Postgres, sets DB_URL,
# and deletes the database afterwards. It never touches your own database.

Output docs/demo.gif

Set Shell "bash"
Set FontSize 15
Set Width 1200
Set Height 720
Set Padding 16
Set TypingSpeed 35ms

# Point gator at the demo database via a temporary HOME, off camera.
Hide
Type `export PATH="$PWD/bin:$PATH" HOME="$(mktemp -d)" PS1='$ '`
Enter
Type `printf '{"db_url":"%s"}' "$DB_URL" > "$HOME/.gatorconfig.json" && clear`
Enter
Show

Type "gator register daniel"
Enter
Sleep 800ms

Type "gator addfeed 'Go Blog' https://go.dev/blog/feed.atom"
Enter
Sleep 800ms

Type "gator addfeed 'Hacker News' https://news.ycombinator.com/rss"
Enter
Sleep 800ms

Type "gator addfeed 'Julia Evans' https://jvns.ca/atom.xml"
Enter
Sleep 800ms

Type "gator feeds"
Enter
Sleep 2s

# Concurrent scrape, then a graceful Ctrl-C.
Type "gator agg --workers 4 1m"
Enter
Sleep 5s
Ctrl+C
Sleep 2s

Type "clear && gator browse 3"
Enter
Sleep 6s
