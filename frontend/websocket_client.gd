extends Node

var socket = WebSocketPeer.new()
var url = "wss://api.kashino.my.id/ws"

func _ready():
	print("Connecting to backend at ", url)
	var err = socket.connect_to_url(url)
	if err != OK:
		print("Could not connect to backend")
		set_process(false)

func _process(_delta):
	socket.poll()
	var state = socket.get_ready_state()
	
	if state == WebSocketPeer.STATE_OPEN:
		while socket.get_available_packet_count() > 0:
			var packet = socket.get_packet()
			var message = packet.get_string_from_utf8()
			print("Received from backend: ", message)
		
		# Send a heartbeat example
		# socket.send_text("Hello from Godot!")
		
	elif state == WebSocketPeer.STATE_CLOSING:
		pass
	elif state == WebSocketPeer.STATE_CLOSED:
		var code = socket.get_close_code()
		var reason = socket.get_close_reason()
		print("WebSocket closed with code: %d, reason %s. Clean: %s" % [code, reason, code != -1])
		set_process(false)
