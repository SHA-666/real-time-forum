let ws = new WebSocket('ws://localhost:8080/ws');
var registrationForm = document.getElementById("registrationForm")
var loginForm = document.getElementById("loginFormu")

function initWebSocket() {
  ws.addEventListener("open", () => {
    console.log("WebSocket connection established");
    var tokenn = localStorage.getItem("Token");
    if (tokenn) {
      showHome();
    }
  });

  ws.addEventListener("error", (error) => {
    console.error("WebSocket error:", error);
  });

  ws.addEventListener("close", () => {
    console.log("WebSocket connection closed");
  });

  function updateTypingLoader(response) {
    // Find the chat box using the data value
    const chatBoxId = `chat-box-${response.data}`;
    const chatBox = document.getElementById(chatBoxId);

    if (chatBox) {
      const typingLoader = chatBox.querySelector('.typing-loader');
      if (response.len >= 1) {
        typingLoader.style.display = 'block';
      } else{
        typingLoader.style.display = 'none';

      }
    }
  }

  ws.onmessage = function (event) {
    var Rdata = JSON.parse(event.data);
    console.log("Data received from backend =", Rdata);
    if (Rdata.datas === "typing" ) {
      updateTypingLoader(Rdata)

    } else if (Rdata.datas === "online" || Rdata.datas === "offline") {
      let userArray = Rdata.data.split(" ");
      updateStatus(Rdata.datas, userArray);
    } else if (Rdata.datas === "AllUsers") {
      afficherUsers(Rdata.data);
    } else if (Rdata.datas === "publications" || Rdata.datas === "userconnection") {
      afficherPublications(Rdata.data);
    } else if (Rdata.datas === "conversation") {
      if (Rdata.data === null) {
        console.log(`panic-unimplemented for the moment (ez to redirect by changin showhomefunc)`);
        return;
      } else {
        Rdata.data.forEach(item => {
          message(
            item.sender_id,
            item.receiver_id,
            item.content,
            item.sendername,
            item.receivername,
            item.sent_at
          );
        });
      }
    } else if (Rdata.datas === "newmessage") {
      message(
        Rdata.data.sender_id,
        Rdata.data.receiver_id,
        Rdata.data.content,
        Rdata.data.sendername,
        Rdata.data.receivername,
        Rdata.data.sent_at
      );
    } else if (Rdata.success) {
      if (Rdata.datas === "register") {
        showLoginForm();
      } else if (Rdata.datas === "login") {
        localStorage.setItem("Token", Rdata.cookies);
        localStorage.setItem("id", Rdata.id);
        showHome();
      }
    } else {
      if (Rdata.datas === "register") {
        document.getElementsByClassName("register-error")[0].textContent = Rdata.error;
      } else if (Rdata.datas === "login") {
        document.getElementsByClassName("login-error")[0].textContent = Rdata.error;
      }
    }
  };
}
function checkAndInitialize() {
  initWebSocket();
  var tokenn = localStorage.getItem("Token");
  if (tokenn) {
    showHome();
  } else {
    showLoginForm();
  }
}
window.onload = checkAndInitialize;
function updateStatus(datas, data) {
  data.forEach(element => {
    let user = document.getElementById("user-" + element);
    if (user) {
      let status = user.querySelector(".status-circle");
      if (datas === "online") {
        status.id = "status-on";
      } else {
        status.id = "status-off";
      }
    }
  });
}
function sendUserId() {
  var userID = localStorage.getItem("id");
  if (!userID) {
    console.error("userID n'a pas été trouvé dans le local storage.");
    return;
  }
  var UserConn = {
    id: userID,
  };
  const Data = {
    data: "userconnection",
    UserConnection: UserConn,
  };

  ws.send(JSON.stringify(Data));

}
function showHome() {
  if (registrationForm) {
    registrationForm.style.display = "none";
  }
  if (loginForm) {
    loginForm.style.display = "none";
  }
  if (ws) {
    sendUserId();
  } else {
    console.error("WebSocket n'est pas prêt pour envoyer des données.(sendUserId)");
  }

  let navbar = document.getElementById("navbar");
  if (navbar) {
    navbar.style.display = "block";
  }
  let creatPost = document.getElementById("creatPost");
  if (creatPost) {
    creatPost.style.display = "block";
  }
  let pubs = document.getElementById("publications-container");
  if (pubs) {
    pubs.style.display = "block";
  }
}
function message(senderId, receiverId, content, senderName, receiverName, timestamp) {
  let str;
  const currentUserId = localStorage.getItem("id");
  if (receiverId === currentUserId) {
    str = senderId;
  } else {
    str = receiverId;
  }

  let chatContent = document.getElementById(`chat-content-${str}`);
  const messageDiv = document.createElement("div");

  // Convert timestamp to a human- eadable date
  const humanDate = new Date(timestamp).toLocaleString();

  if (senderId !== currentUserId) {
    messageDiv.className = "message message-receiver";
    messageDiv.innerHTML = `<strong>${senderName}</strong> <em>${humanDate}</em>: ${content}`;
  } else {
    messageDiv.className = "message message-sender";
    messageDiv.innerHTML = `<strong>${senderName}</strong> <em>${humanDate}</em>: ${content}`;
  }

  chatContent.appendChild(messageDiv);
}
function showmessage() {
  let users = document.getElementById("users")
  if (users.style.display === "block") {
    users.style.display = "none"
  } else {
    users.style.display = "block"
  }
}
function afficherPublications(publications) {
  const fragment = document.createDocumentFragment();
  const publicationsContainer = document.getElementById('publications-container');
  publicationsContainer.innerHTML = '';
  publications.forEach(publication => {
    const publicationElement = document.createElement('div');
    publicationElement.classList.add('publication');

    const titleElement = document.createElement('h2');
    titleElement.textContent = publication.title;

    const categoriesElement = document.createElement('p');
    categoriesElement.textContent = `Catégories: ${publication.categories}`;

    const contentElement = document.createElement('p');
    contentElement.textContent = publication.content;

    const creationDateElement = document.createElement('p');
    const humanDate = new Date(publication.creation_date).toLocaleString();
    creationDateElement.textContent = `Date de création: ${humanDate}`;

    const userElement = document.createElement('p');
    userElement.textContent = `Publié par: ${publication.user}`;

    const showCommentsButton = document.createElement('button');
    showCommentsButton.textContent = 'Voir les commentaires';
    showCommentsButton.classList.add('comment-button');

    showCommentsButton.addEventListener('click', () => {
      const commentsSection = publicationElement.querySelector('.comments-section');
      const commentForm = publicationElement.querySelector('.comment-form');
      if (commentsSection && commentForm) {
        const isHidden = commentsSection.style.display === 'none';
        commentsSection.style.display = isHidden ? 'block' : 'none';
        commentForm.style.display = isHidden ? 'block' : 'none';
      } else {
        if (!commentsSection) {
          afficherComments(publicationElement, publication);
        }
        if (!commentForm) {
          afficherInputCommentaires(publicationElement, publication);
        }
      }
    });

    publicationElement.appendChild(titleElement);
    publicationElement.appendChild(userElement);
    publicationElement.appendChild(categoriesElement);
    publicationElement.appendChild(contentElement);
    publicationElement.appendChild(creationDateElement);
    publicationElement.appendChild(showCommentsButton);

    fragment.appendChild(publicationElement);
  });

  publicationsContainer.appendChild(fragment);
}
function afficherComments(publicationElement, publication) {
  console.log("Rendering comments for publication:", publication);

  let commentsSection = publicationElement.querySelector('.comments-section');

  if (!commentsSection) {
    commentsSection = document.createElement('div');
    commentsSection.classList.add('comments-section');
  } else {
    commentsSection.innerHTML = ''; // Clear existing comments
  }

  if (publication.comments && publication.comments.length > 0) {
    publication.comments.forEach(comment => {
      console.log("Rendering comment:", comment);

      const commentDiv = document.createElement('div');
      commentDiv.classList.add('comment');

      const username = document.createElement('span');
      username.textContent = comment.username;

      const content = document.createElement('p');
      content.textContent = comment.content;

      const creationDate = document.createElement('span');
      const humanDate = new Date(comment.creation_date).toLocaleString();
      creationDate.textContent = humanDate;

      commentDiv.appendChild(username);
      commentDiv.appendChild(content);
      commentDiv.appendChild(creationDate);

      commentsSection.appendChild(commentDiv);
    });
  } else {
    const noCommentsMessage = document.createElement('p');
    noCommentsMessage.textContent = 'Aucun commentaire pour le moment.';
    commentsSection.appendChild(noCommentsMessage);
  }

  publicationElement.appendChild(commentsSection);
  afficherInputCommentaires(publicationElement, publication);
}

function afficherInputCommentaires(publicationElement, publication) {
  let commentForm = publicationElement.querySelector('.comment-form');

  if (!commentForm) {
    commentForm = document.createElement('form');
    commentForm.classList.add('comment-form');

    const input = document.createElement('input');
    input.setAttribute('type', 'text');
    input.setAttribute('placeholder', 'Entrez votre commentaire');
    input.classList.add('comment-input');
    commentForm.appendChild(input);

    const sendButton = document.createElement('button');
    sendButton.textContent = 'Envoyer';
    sendButton.classList.add('send-button');
    sendButton.type = 'button'; // Change the type to "button" to prevent form submission
    commentForm.appendChild(sendButton);

    sendButton.addEventListener("click", function (event) {
      event.preventDefault();
      let comment = {
        user: localStorage.getItem("id"),
        comment: input.value,
        idpub: publication.idpub,
      };
      // Retrieve form data
      const formData = {
        data: "commentaire",
        commentaire: comment,
      };
      ws.send(JSON.stringify(formData));
      console.log("Form sent:", formData);
      input.value = "";
    });

    publicationElement.appendChild(commentForm);
  }
}
function afficherUsers(params) {
  var userList = document.getElementById("users");
  userList.innerHTML = "";
  var id = localStorage.getItem("id");

  params.forEach(function (user) {
    let userDiv = document.createElement("div");
    userDiv.className = "user";
    userDiv.id = "user-" + user.Id;

    let userName = document.createElement("span");
    userName.className = "nickname";
    userName.textContent = user.Name;

    let statusCircle = document.createElement("div");
    statusCircle.className = "status-circle";
    statusCircle.id = "status-off";

    userDiv.addEventListener("click", function () {
      let chatbx = document.querySelectorAll(".chat-box");
      chatbx.forEach(element => {
        element.style.display = "none";
      });

      let existingChatBox = document.getElementById("chat-box-" + user.Id);
      if (existingChatBox) {
        existingChatBox.style.display = "block";
      } else {
        let chatBox = document.createElement("div");
        chatBox.className = "chat-box";
        chatBox.id = "chat-box-" + user.Id;
        chatBox.style.display = "block";

        const chatHeader = document.createElement("h2");
        chatHeader.textContent = "Chat with " + user.Name;
        const chatContent = document.createElement("div");
        chatContent.id = "chat-content-" + user.Id;

        const typingLoader = document.createElement("div");
        typingLoader.className = "typing-loader";
        typingLoader.style.display = "none";


        const chatInput = document.createElement("input");
        chatInput.type = "text";
        chatInput.placeholder = "Écrivez votre message...";

        const sendButton = document.createElement("button");
        sendButton.textContent = "Send";



        const message = {
          data: "getmessages",
          MessageInfo: {
            sender_id: localStorage.getItem("id"),
            receiver_id: chatBox.id,
            content: " ",
          },
        };
        ws.send(JSON.stringify(message));

        chatInput.addEventListener("input", function () {
          const inoutmss = chatInput.value;
          const lenofchat = inoutmss.length;

          const typingMessage = {
            data: "typing",
            typing: {
              sender_id: localStorage.getItem("id"),
              receiver_id: chatBox.id,
              len : lenofchat
            }
          };
          console.log("zab$$$$$$$$$$$$$$$$$$a",lenofchat)
          ws.send(JSON.stringify(typingMessage));
        });

        sendButton.addEventListener("click", function () {
          const messageContent = chatInput.value.trim();
          if (messageContent) {
            const newMessage = {
              data: "newmessages",
              MessageInfo: {
                sender_id: localStorage.getItem("id"),
                receiver_id: chatBox.id,
                content: messageContent,
              },
            };
            ws.send(JSON.stringify(newMessage));
            chatInput.value = "";
          }
        });

        chatBox.appendChild(chatHeader);
        chatBox.appendChild(chatContent);
        chatBox.appendChild(typingLoader);
        chatBox.appendChild(chatInput);
        chatBox.appendChild(sendButton);
        userList.appendChild(chatBox);
      }
    });

    userDiv.appendChild(userName);
    userDiv.appendChild(statusCircle);
    userList.appendChild(userDiv);
  });
}
//event listener
document.getElementById("PostCreatBTN").addEventListener("click", function (event) {
  // Prevent default form submission behavior
  event.preventDefault();
  var creatPost = {
    id: localStorage.getItem("id"),
    title: document.getElementById("postTitle").value,
    category: document.getElementById("postCategory").value,
    content: document.getElementById("postContent").value,

  }
  // Retrieve form data
  const formData = {
    data: "creatPost",
    CreatPostData: creatPost,
  };

  let errorMessage = document.getElementById("error-message");
  if (creatPost.title === "" || creatPost.content === "") {
    errorMessage.textContent = "Veuillez remplir tous les champs.";
    return
  } else {
    ws.send(JSON.stringify(formData));
  }
  document.getElementById("postForm").style.display = "none"
  document.getElementById("publications-container").style.display = "block"
});
let postBTN = document.getElementById("createPostBtn")
postBTN.addEventListener("click", function (event) {
  let pubs = document.getElementById("publications-container")
  let postForm = document.getElementById("postForm")
  if (postForm.style.display === "block") {
    postForm.style.display = "none"
    pubs.style.display = "block"
  } else {
    postForm.style.display = "block"
    pubs.style.display = "none"
  }
  if (postBTN.innerHTML == "voir les publications") {
    postBTN.innerHTML = "créer une publication"
  } else {
    postBTN.innerHTML = "voir les publications"
  }


})
document.getElementById("loginBTN").addEventListener("click", function (event) {
  // Prevent default form submission behavior
  event.preventDefault();
  let login = {
    nickname: document.getElementById("loginUsername").value,
    password: document.getElementById("loginPassword").value
  }
  // Retrieve form data
  const formData = {
    data: "login",
    loginData: login,
  };
  let errorMessage = document.getElementById("error-messageL");

  if (login.nickname === "" || login.password === "") {
    errorMessage.textContent = "Veuillez remplir tous les champs.";
    return
  } else {
    ws.send(JSON.stringify(formData));
  }
  console.log("formulaire login sent from frontend to back end =", formData)
});
document.getElementById("registerBTN").addEventListener("click", function (event) {
  event.preventDefault();
  let register = {
    nickname: document.getElementById("nickname").value,
    age: parseInt(document.getElementById("age").value),
    gender: document.getElementById("gender").value,
    first_name: document.getElementById("first_name").value,
    last_name: document.getElementById("last_name").value,
    email: document.getElementById("email").value,
    password: document.getElementById("password").value
  }
  const formData = {
    data: "register",
    registerData: register,
  };
  let errorMessage = document.getElementById("error-message");
  if (register.nickname === "" || register.age === "" || register.first_name === "" || register.last_name === "" || register.email === "" || register.password === "") {
    errorMessage.textContent = "Veuillez remplir tous les champs.";
    return
  } else {
    ws.send(JSON.stringify(formData));
  }
  console.log("formulaire register sent from frontend to back end =", formData)
})
document.getElementById("navmessage").addEventListener("click", function (event) {
  let publication = document.getElementById("publications-container")
  let creatPost = document.getElementById("creatPost")
  if (creatPost.style.display === "block") {
    users.style.display = "block"
    creatPost.style.display = "none"
    publication.style.display = "none"
  } else {
    creatPost.style.display = "block"
    publication.style.display = "block"
    users.style.display = "none"
  }
})

//onclick function
function logout() {
  localStorage.removeItem("Token");
  localStorage.removeItem("id");
  // localStorage.clear()
  location.reload();
}
function requestinfo() {
  let id = localStorage.getItem("id")
  const formData = {
    data: "profil",
    IdOfUser: { UserId: id },
  };
  ws.send(JSON.stringify(formData));
}
function showRegistrationForm() {
  registrationForm.style.display = "block";
  loginForm.style.display = "none";
}
function showLoginForm() {
  loginForm.style.display = "block";
  registrationForm.style.display = "none";
}
