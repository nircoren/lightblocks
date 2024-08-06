Quick Start:
  init:
    clone repo https://github.com/nircoren/lightblocks.git
    put .env on root dir
    docker-compose up --build
  Send messages:
    $ docker exec <containter_name> /app/client --username <username> --msgs <msgs>
    Done!
    
    You need to pass user name and message in this format:
    username string
    type Message struct {
      Command string `json:"command"`
      Key     string `json:"key,omitempty"`
      Value   string `json:"value,omitempty"`
    }

    example: docker exec lightblocks-client-1 /app/client --username nir --msgs '[{"command":"addItem","key":"key1","value":"value1"},{"command":"addItem","key":"key2","value":"value2"},{"command":"addItem","key":"key3","value":"value3"},{"command":"addItem","key":"key111","value":"yaythere"},{"command":"getAllItems"}]'
    example multiple users: [example_multiple_users.txt](https://github.com/user-attachments/files/16509687/example_multiple_users.txt)

Assumptions:
	Don't need to remove command from map after making it.
	You didn't want client to be a server. To make it accessible with docker I made an infinite loop.
	Order of actions should stay the same for each client, but doesnt matter between 2 clients.
	Shouldn'tt cancel sending and receiveing based on one bad message.
	Shouldn't remove from orederedMap the action after execute.


sequence diagram: ![image](https://github.com/user-attachments/assets/6576bc41-03c6-4500-ba8e-e94ea800a2f6)
