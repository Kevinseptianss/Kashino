extends Node

signal login_success(user_data)
signal login_failed(error_message)
signal signup_success(user_data)
signal signup_failed(error_message)
signal balance_updated(new_balance)
signal history_received(history)
signal connection_status_changed(online)
signal rooms_received(rooms)
signal room_joined(room_data)
signal player_sat(room_data)
signal player_stood_up()
signal room_update(room_data)

var base_url = "http://localhost:9090"
var current_user = null
var current_balance = 0.0
var socket = WebSocketPeer.new()
var url = "ws://localhost:9090/ws"
var ws_connected = false
var auth_file = "user://auth.cfg"
var _last_login_data = {"u":"","p":""}
var is_side_by_side = false
var instance_count = 1
var instance_index = 0

func _ready():
	process_mode = PROCESS_MODE_ALWAYS
	
	# Support multiple profiles for local testing
	# Usage: godot -- --profile=player1 --side-by-side=2
	for arg in OS.get_cmdline_user_args():
		if arg.begins_with("--profile="):
			var profile_name = arg.split("=")[1]
			instance_index = int(profile_name) - 1 # Assuming numeric profiles for positioning
			auth_file = "user://auth_" + profile_name + ".cfg"
			print("AuthService: Using profile: ", profile_name, " (", auth_file, ")")
		elif arg.begins_with("--side-by-side="):
			is_side_by_side = true
			instance_count = int(arg.split("=")[1])
			print("AuthService: Side-by-side mode enabled (Count: ", instance_count, ")")

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
			
			# Handle multiple messages in one packet (concatenated with \n)
			# We use a loop to ensure we catch all messages in the buffer
			var messages = message.split("\n", false)
			for msg in messages:
				var trimmed = msg.strip_edges()
				if trimmed == "": continue
				# If somehow two JSONs are still stuck together (no \n), we catch it in handle
				_handle_ws_message(trimmed)
			
	elif state == WebSocketPeer.STATE_CLOSED or state == WebSocketPeer.STATE_CLOSING:
		if ws_connected:
			_set_online(false)
			# Do not auto-reconnect if not logged in
			if current_user:
				print("AuthService: Connection lost, retrying in 2 seconds...")
				get_tree().create_timer(2.0).timeout.connect(_connect_to_ws)

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
	_last_login_data = {"u":username, "p":password}
	
	var headers = ["Content-Type: application/json"]
	var err = http_request.request(base_url + "/signin", headers, HTTPClient.METHOD_POST, body)
	if err != OK:
		login_failed.emit("Could not make login request")

func signup(username, email, password, captcha_answer, captcha_expected):
	var http_request = HTTPRequest.new()
	add_child(http_request)
	http_request.request_completed.connect(_on_signup_request_completed)
	
	var body = JSON.stringify({
		"username": username,
		"email": email,
		"password": password,
		"captcha_answer": captcha_answer,
		"captcha_expected": captcha_expected
	})
	
	var headers = ["Content-Type: application/json"]
	var err = http_request.request(base_url + "/signup", headers, HTTPClient.METHOD_POST, body)
	if err != OK:
		signup_failed.emit("Could not make signup request")

func _on_login_request_completed(_result, response_code, _headers, body):
	var response = JSON.parse_string(body.get_string_from_utf8())
	if response_code == 200:
		current_user = {
			"id": str(response["id"]),
			"username": response["username"],
			"balance": response["balance"],
			"tier": response.get("tier", "VIP Silver")
		}
		current_balance = current_user["balance"]
		login_success.emit(current_user)
		save_auth(_last_login_data["u"], _last_login_data["p"])
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

func get_rooms():
	print("AuthService: Requesting rooms...")
	_send_ws_message("get_rooms", {})

func update_balance(amount: float, source: String):
	_send_ws_message("update_balance", {
		"amount": amount,
		"source": source
	})

func _handle_ws_message(json_str):
	print("RAW WS RECEIVED: ", json_str)
	var response = JSON.parse_string(json_str)
	if not response: 
		print("Failed to parse WS message: ", json_str)
		return
	
	var action = response.get("action", "")
	var status = response.get("status", "")
	var data = response.get("data", {})
	
	print("Parsed Action: ", action, " Status: ", status)
	
	match action:
		"get_balance", "balance_update":
			if status == "success":
				current_balance = data["balance"]
				balance_updated.emit(current_balance)
		"get_history":
			if status == "success":
				history_received.emit(data["history"])
		"get_rooms":
			if status == "success":
				print("AuthService: Received rooms, emitting signal. Data: ", data)
				rooms_received.emit(data)
			else:
				print("AuthService: get_rooms FAILED. Status: ", status)
		"join_room":
			if status == "success":
				room_joined.emit(data)
		"player_sat":
			if status == "success":
				player_sat.emit(data)
		"room_update":
			if status == "success":
				room_update.emit(data)
		"standup":
			if status == "success":
				player_stood_up.emit()

func _send_ws_message(action, data):
	if socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		print("WS SEND FAIL: Not connected (State: ", socket.get_ready_state(), ")")
		return
		
	var msg = JSON.stringify({
		"action": action,
		"data": data
	})
	var err = socket.send_text(msg)
	if err != OK:
		print("WS SEND ERROR: ", err, " for action: ", action)
	else:
		print("WS SENT: ", action)

func save_auth(username, password):
	var config = ConfigFile.new()
	config.set_value("auth", "username", username)
	config.set_value("auth", "password", password)
	config.save(auth_file)

func load_auth():
	var config = ConfigFile.new()
	var err = config.load(auth_file)
	if err != OK: return null
	return {
		"username": config.get_value("auth", "username", ""),
		"password": config.get_value("auth", "password", "")
	}

func clear_auth():
	var fname = auth_file.replace("user://", "")
	var dir = DirAccess.open("user://")
	if dir.file_exists(fname):
		dir.remove(fname)
	current_user = null
	ws_connected = false
	socket.close()

func auto_login():
	var creds = load_auth()
	if creds and creds["username"] != "" and creds["password"] != "":
		print("Auto-logging in user: ", creds["username"])
		login(creds["username"], creds["password"])
