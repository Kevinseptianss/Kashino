extends Node

signal login_success(user_data)
signal login_failed(error_message)
signal signup_success(user_data)
signal signup_failed(error_message)
signal balance_updated(new_balance)
signal history_received(history)
signal connection_status_changed(online)

var base_url = "http://localhost:9090"
var current_user = null
var current_balance = 0.0
var socket = WebSocketPeer.new()
var url = "ws://localhost:9090/ws"
var ws_connected = false

func _ready():
	process_mode = PROCESS_MODE_ALWAYS

func _connect_to_ws():
	var ws_url = url + "?user_id=" + str(current_user["id"])
	print("Connecting to backend at ", ws_url)
	var err = socket.connect_to_url(ws_url)
	if err != OK:
		print("Could not connect to backend")
		_set_online(false)

func _process(_delta):
	socket.poll()
	var state = socket.get_ready_state()
	
	if state == WebSocketPeer.STATE_OPEN:
		if not ws_connected:
			_set_online(true)
		
		while socket.get_available_packet_count() > 0:
			var packet = socket.get_packet()
			var message = packet.get_string_from_utf8()
			_handle_ws_message(message)
			
	elif state == WebSocketPeer.STATE_CLOSED or state == WebSocketPeer.STATE_CLOSING:
		if ws_connected:
			_set_online(false)
			# Do not auto-reconnect if not logged in
			if current_user:
				_connect_to_ws()

func _set_online(online):
	ws_connected = online
	connection_status_changed.emit(online)

func login(username, password):
	var http_request = HTTPRequest.new()
	add_child(http_request)
	http_request.request_completed.connect(_on_login_request_completed)
	
	var body = JSON.stringify({
		"username": username,
		"password": password
	})
	
	var headers = ["Content-Type: application/json"]
	var err = http_request.request(base_url + "/signin", headers, HTTPClient.METHOD_POST, body)
	if err != OK:
		login_failed.emit("Could not make login request")

func signup(username, email, password):
	var http_request = HTTPRequest.new()
	add_child(http_request)
	http_request.request_completed.connect(_on_signup_request_completed)
	
	var body = JSON.stringify({
		"username": username,
		"email": email,
		"password": password
	})
	
	var headers = ["Content-Type: application/json"]
	var err = http_request.request(base_url + "/signup", headers, HTTPClient.METHOD_POST, body)
	if err != OK:
		signup_failed.emit("Could not make signup request")

func _on_login_request_completed(_result, response_code, _headers, body):
	var response = JSON.parse_string(body.get_string_from_utf8())
	if response_code == 200:
		current_user = {
			"id": response["id"],
			"username": response["username"],
			"balance": response["balance"]
		}
		current_balance = current_user["balance"]
		login_success.emit(current_user)
		# Connect to WS after successful login
		_connect_to_ws()
	else:
		var error = "Login failed"
		if response and response.has("error"):
			error = response["error"]
		login_failed.emit(error)

func _on_signup_request_completed(_result, response_code, _headers, body):
	var response = JSON.parse_string(body.get_string_from_utf8())
	if response_code == 201:
		signup_success.emit(response)
	else:
		signup_failed.emit("Signup failed")

func fetch_balance():
	_send_ws_message("get_balance", {})

func get_history():
	_send_ws_message("get_history", {})

func _handle_ws_message(json_str):
	var response = JSON.parse_string(json_str)
	if not response: return
	
	var action = response.get("action", "")
	var status = response.get("status", "")
	var data = response.get("data", {})
	
	match action:
		"get_balance":
			if status == "success":
				current_balance = data["balance"]
				balance_updated.emit(current_balance)
		"get_history":
			if status == "success":
				history_received.emit(data["history"])

func _send_ws_message(action, data):
	if socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		print("Cannot send message, WebSocket not connected")
		return
		
	var msg = JSON.stringify({
		"action": action,
		"data": data
	})
	socket.send_text(msg)
