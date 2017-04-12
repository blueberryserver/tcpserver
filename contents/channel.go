package contents

// channel
type Channel struct {
	cNo uint32 // channel number

	members map[uint32]User // channel user
}

// enter channel
// leave channel
