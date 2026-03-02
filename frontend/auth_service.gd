extends Node

signal login_success(user_data)
signal login_failed(error_message)
signal signup_success(user_data)
signal signup_failed(error_message)
signal balance_updated(new_balance)

var base_url = "http://localhost:9090"
var current_user = null
var current_balance = 0.0
var auth_token = ""

func _ready():
	process_mode = PROCESS_MODE_ALWAYS

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

func fetch_balance(user_id):
	var http_request = HTTPRequest.new()
	add_child(http_request)
	http_request.request_completed.connect(_on_balance_request_completed)
	
	var headers = ["Authorization: Bearer " + auth_token]
	var err = http_request.request(base_url + "/balance?id=" + user_id, headers)
	if err != OK:
		print("Could not fetch balance")

func _on_login_request_completed(result, response_code, headers, body):
	var response = JSON.parse_string(body.get_string_from_utf8())
	if response_code == 200:
		auth_token = response["token"]
		current_user = {
			"id": response["id"],
			"username": response["username"],
			"balance": response["balance"]
		}
		current_balance = current_user["balance"]
		login_success.emit(current_user)
	else:
		var error = "Login failed"
		if response and response.has("error"):
			error = response["error"]
		login_failed.emit(error)

func _on_signup_request_completed(result, response_code, headers, body):
	var response = JSON.parse_string(body.get_string_from_utf8())
	if response_code == 201:
		signup_success.emit(response)
	else:
		signup_failed.emit("Signup failed")

func _on_balance_request_completed(result, response_code, headers, body):
	var response = JSON.parse_string(body.get_string_from_utf8())
	if response_code == 200:
		current_balance = response["balance"]
		balance_updated.emit(current_balance)
