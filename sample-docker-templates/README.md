# Sample Docker Templates

This directory contains sample Dockerfile templates for different languages and
frameworks which can be used for further containerizing the application and deploy
it in production environment. The template contains a generic approach for building
containerized application with server config files included in it.

## Contributing Guidelines

**Each and every contribution matters**

You are most welcome to contribute a sample Dockerfile template for the languages or
frameworks not included yet. Feel free to contribute Dockerfile templates keeping
in mind the following Guidelines - 

* The Dockerfile should be a generic template for the language or framework chosen
* Directory Structure 

	```bash
	├── flask   			# dir should be framework name (lower-case)
	│	├── Dockerfile		# Dockerfile
	│	├── nginx.default	# server files
	│	├── start.sh		# shell scripts
	│	└── uwsgi.ini
	├── go				# dir should be language name (lower-case)
	│	└── Dockerfile		# Dockerfile
	├── node

	```
	
* Include appropriate comments in Dockerfile

## Communications

You are most welcome to the community. 

You can join us at **[discord](https://discord.gg/jsRG5qx2gp)**  and use the __#contrib__ channel
to start contributing and solve your all doubts.

