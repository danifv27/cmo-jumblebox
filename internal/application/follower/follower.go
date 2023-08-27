package follower

type Follower interface {
	// Lines continuously emits a stream of lines
	Lines() chan string
}
