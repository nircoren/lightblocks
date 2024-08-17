<h2>Updated Project: </h2>
<h3>Quick Start</h3>:
	1. run in cmd: $ git clone https://github.com/nircoren/lightblocks.git <br />
	2. put .env i sent you on root dir. make sure its called .env <br />
	3. build images: $ docker-compose build <br />
	4. run server: $ docker-compose up -d server <br />
	5. run client (remove container after finish): $ docker run --rm -it lightblocks-client <br />
	6. first prompt will be username, second should be a json in this format: <br />
		[
			{
				"Action": "addItem",
				"Key": "1",
				"Value": "val1"
			},
			{
				"Action": "getAllItems"
			},
			{
				"Action": "addItem",
				"Key": "2",
				"Value": "val2"
			}
		]
<br/><br/>
  !!! might have issue with parsing the json on bash/wt in windows, better to use another terminal. <br />
  
Example multiple inputs:  <br/>
[input1.json](https://github.com/user-attachments/files/16645653/input1.json)  <br/>
[input2.json](https://github.com/user-attachments/files/16645651/input2.json)  <br/>
[input3.json](https://github.com/user-attachments/files/16645652/input3.json)  <br/>



You can try to run multiple instances of the program with different username for each input



<h3> My assumptions during project: </h3>
	Don't need to remove command from map execute making it. <br />
	Client shouldn't be a server. <br />
	Order of actions should stay the same for each client, but doesnt matter if order is not the same between 2 different clients <br />
	Shouldn't cancel sending and receiveing based on one bad message. <br />

