package piedpiper

// TODO: Abstract away each pipeline stage worker calls
type Pipe interface {
	Process(in chan *Ad) chan *Ad
}

func NewPipeline(pipes ...Pipe) Pipeline {
	head := make(chan *Ad, 10000)
	var nextChannel chan *Ad
	for _, pipe := range pipes {
		if nextChannel == nil {
			nextChannel = pipe.Process(head)
		} else {
			nextChannel = pipe.Process(nextChannel)
		}
	}
	return Pipeline{head: head, tail: nextChannel}
}

type Pipeline struct {
	head chan *Ad
	tail chan *Ad
}

func (p *Pipeline) Enqueue(item *Ad) {
	p.head <- item
}

func (p *Pipeline) Dequeue(handler func(*Ad)) {
	for i := range p.tail {
		handler(i)
	}
}

func (p *Pipeline) Close() {
	close(p.head)
}
