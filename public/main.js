const socket = io();

new Vue({
  el: '#app',
  created() {

    //This funtion handles chat messages to display them in the chat room
    //Prevents unintended users from watching private messages
    socket.on("chat message", (message, user, recv, pmsg, usersRoom) => {
      if(pmsg)
      {
        if(user != this.user){

          if(recv != this.user)
            return
        }
      }
      this.messages.push({
        text: message,
        date: `${new Date().getHours()}:${new Date().getMinutes()}`,
        user: user
      })
      this.users = []
      usersRoom.forEach(user => {
        this.adduser({username:user})         
      })
    })

    //User has been subscribed and can use chat room    
    socket.on("logged", (user, loggedUsers) => {
        this.joined = true
        this.user = user
        loggedUsers.forEach(user => {
          this.adduser({username:user})         
        })
      })

      //Display notice if unsubscribed user tries to login 
      socket.on("unknown_user", (message) => {
          alert("Please signIn before login");
      })

      //Display notice if user entered invalid data in login form
      socket.on("invalid_credentials", (message) => {
          alert("Invalid username or password. Please try again\nOr please sign in");
      })

      //Display notice after user successfully signed in
      socket.on("signed_user", (message) => {
          alert("You can login now!");
      })

      //Display notice if trying to signin with another user's username
      socket.on("used_username", function (){
          alert("Username already in use by another user");
      })

      //Update user presence in the chat room
      socket.on("in_session", (user) => {
        this.joined = true
        this.user = user
      })

      //Update users list when user leaves chat room
      socket.on("user_removed", (loggedUsers) => {
        this.joined = false
        this.user = ""
        this.users = []
        loggedUsers.forEach(user => {
          this.adduser({username:user})         
        })
        socket.emit("user_left")
      })

      //Notice user in case of database error
      socket.on("error_signin", () =>{
        console.log("Unable to subscribe user. Try again!")
      })   
        
  },
  data: {
    message: '',
    messages: [],
    joined: false,
    password: null, 
    username: null, 
    user: '',
    users: [],
    search: '',
    pmsg: false,
    receiver: 'any'
  },
  methods: {

    //Function to trigger message display in chat room (either public or private messages)
    sendMessage() {
      if(!this.message)
      {
        return
      }
       
      socket.emit("chat message", this.message, this.user, this.receiver, this.pmsg)
      this.message = "";
    },

    //Function to send client data for user authentication against server side database
    login: function () {
        if (!this.username) {
            alert('You must enter a username');
            return
        }
        if (!this.password) {
            alert('You must enter a password');
            return
        }
        socket.emit("authentication", JSON.stringify({
            user: this.username, 
            password: this.password}))
    },

    //Function to send client data for user registration in server side database
    signin: function () {
        if (!this.username) {
            alert('You must enter a username');
            return
        }
        if (!this.password) {
            alert('You must enter a password');
            return
        }
        socket.emit("signing", JSON.stringify({
          user: this.username, 
          password: this.password}))
    },

    //Function to add a client to users list
    adduser: function (user) {
        this.users.push(user);
    },

    //Function to anounce a client is leaving the chat room
    leave: function(){
      alert('You are leaving this conversation');
      socket.emit("leave_room", this.user)
    },

    // Function to send a private message to a single user
    privatemsg: function(receiver){
      this.pmsg = false
      this.receiver = 'any'

      if (receiver != this.user){
        this.receiver = receiver
        this.pmsg = true
      }      
    },   
  },

  computed: {
    //Function to filter users in the search box
    filteredList() {
      return this.users.filter(user => {
        return user.username.toLowerCase().includes(this.search.toLowerCase())
      })
    }
  }
})