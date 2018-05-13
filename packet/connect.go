package packet

type ConnectFlags struct {
	UserName     bool
	Passwore     bool
	WillRetain   bool
	WillQoS      byte // 2 bytes actually
	WillFlag     bool
	CleanSession bool
}

type VariableHeaderConnect struct {
	ProtocolName  string
	ProtocolLevel byte
	ConnectFlags  ConnectFlags
	KeepAlive     int
}

type ConnectControlPacket struct {
	FixedHeader    FixedHeader
	VariableHeader VariableHeaderConnect

	ClientIdentifier string
}
