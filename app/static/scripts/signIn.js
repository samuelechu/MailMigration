function setElements(isLoggedIn){
  var auth2 = gapi.auth2.getAuthInstance();
  var googleUser= auth2.currentUser.get();
  if(isLoggedIn){
      document.getElementById('gSignInButton').style.display = 'none';
      document.getElementById('logout').style.display = 'block';

      var profile = googleUser.getBasicProfile()
      var uid = profile.getId() //safe to use because user token was checked in onSignIn()

      //if there is no job in progress, show select Section
      jobInProgress(uid, manageSections);   

  } else {
      $("#finishSection").collapse('hide');
      $("#jobSection").collapse('hide');
      $("#transferButtonSection").collapse('hide');
      $("#selectSection").collapse('hide');
      document.getElementById('logout').style.display = 'none';
      document.getElementById('gSignInButton').style.display = 'block';
  }
}

function signOut() {
  var auth2 = gapi.auth2.getAuthInstance();

  var xhr = new XMLHttpRequest();
  xhr.open('GET', 'deleteCookies');
  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
  xhr.onload = function() {
    location.reload();
    auth2.signOut().then(function () {
      setElements(false);
      console.log('User signed out.');
    });
  };

  xhr.send();
}

function sendTokentoDB(id_token){
  var auth2 = gapi.auth2.getAuthInstance();
  var googleUser= auth2.currentUser.get();
  var profile = googleUser.getBasicProfile();
  
  var xhr = new XMLHttpRequest();
  xhr.open('POST', 'signIn');
  xhr.setRequestHeader("Content-Type", "application/json");
  xhr.onload = function() {
    setElements(true);
  };

  var data = {
      Uid : id_token
    , Name : profile.getName()
  }

  xhr.send(JSON.stringify(data));

  console.log("Sent: " + JSON.stringify(data) + " to database");

}

//after sign in, verify token
function onSignIn(googleUser) {
  document.getElementById('gSignInButton').style.display = 'none';
  var id_token = googleUser.getAuthResponse().id_token;

  var xhr = new XMLHttpRequest();
  xhr.open('GET', 'https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=' + id_token);
  xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
  xhr.onload = function() {
    console.log('Signed in as: ' + xhr.responseText);
    var resp = JSON.parse(xhr.responseText);

    //token is valid, send to back end
    if (resp.aud === "65587295914-kbl4e2chuddg9ml7d72f6opqhddl62fv.apps.googleusercontent.com") {
      sendTokentoDB(resp.sub);
    } else {
      signOut();
    }
  };
  xhr.send();
  

  var profile = googleUser.getBasicProfile();
  console.log('ID: ' + profile.getId()); // Do not send to your backend! Use an ID token instead.
  console.log('Name: ' +  '{{.UserName}}');//profile.getName());
  console.log('Image URL: ' + profile.getImageUrl());
  console.log('Email: ' + profile.getEmail()); // This is null if the 'email' scope is not present.
  console.log('id_token: ' + id_token);
  
}
