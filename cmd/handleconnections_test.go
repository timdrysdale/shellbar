package cmd

// TODO refactor with interface in client struct to allow mocking
// of *websocket.Conn https://github.com/gorilla/websocket/issues/74
// that will allow readpump to be tested on its own.
// For now, see integration test instead.
