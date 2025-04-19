package election

const (
	// EventLeaderChanged is the event when other node becomes the leader.
	EventLeaderChanged string = "leader_changed"
	// EventBecomeLeader is the event when the current node becomes the leader.
	EventBecomeLeader string = "become_leader"
	// EventLoseLeader is the event when the current node loses the leader.
	EventLoseLeader string = "lose_leader"
)
