# livecoding-tv-friend-finder
A small command-line app for finding your buddy on https://livecoding.tv.

What it does:
- 1) Get current live channels
- 2) Connect to XMPP chat given your credentials
- 3) Fetch all users in live channels
- 4) Compare with `<username-to-find>` given to app
- 5) Send back channels where `<username-to-find>` was found

### Installation
```
$ go get -u github.com/ThomasBS/livecoding-tv-friend-finder
```

### Usage
You need to find your password to the livecoding.tv XMPP chat. This can be done by:
- 1) Open some live stream page when logged in
- 2) Open dev tools and go to elements tab
- 3) Search the HTML for "password"
- 4) The XMPP password will be a long string in an object containing jid and password

Once equipped with your password, you can run this app with:

```
$ livecoding-tv-friend-finder <username-to-find> <your-username> <your-password>
```

i.e:

```
$ livecoding-tv-friend-finder baz thomasbs n12k4bjebberish3u1n
```
