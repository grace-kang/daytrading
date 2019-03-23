$(document).ready(function() {
  $("#myButtons :input").change(function() {
    console.log("button clicked");
    console.log(this); // points to the clicked input button
  });
});
