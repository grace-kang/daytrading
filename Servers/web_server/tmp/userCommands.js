var username;

$(document).ready(function() {
  $("#myButtons :input").change(function() {
    console.log("id is " + this.id);
    currentCommand = this.id;
    switch (currentCommand) {
      case "ADD":
        $("#fieldOne").show();
        $("#fieldTwo").hide();
        $("#fieldOneLabel").text("Amount");
        break;
      case "QUOTE":
        $("#fieldOne").hide();
        $("#fieldTwo").show();
        $("#fieldTwoLabel").text("Stock Symbol");
        break;
      case "BUY":
        $("#fieldOne").show();
        $("#fieldTwo").show();
        $("#fieldOneLabel").text("Amount");
        $("#fieldTwoLabel").text("Stock Symbol");
        break;
      case "COMMIT_BUY":
        $("#fieldOne").hide();
        $("#fieldTwo").hide();
        break;
      case "CANCEL_BUY":
        $("#fieldOne").hide();
        $("#fieldTwo").hide();
        break;
      case "SELL":
        $("#fieldOne").show();
        $("#fieldTwo").show();
        $("#fieldOneLabel").text("Amount");
        $("#fieldTwoLabel").text("Stock Symbol");
        break;
      case "COMMIT_SELL":
        $("#fieldOne").hide();
        $("#fieldTwo").hide();
        break;
      case "CANCEL_SELL":
        $("#fieldOne").hide();
        $("#fieldTwo").hide();
        break;
      case "SET_BUY_AMOUNT":
        $("#fieldOne").show();
        $("#fieldTwo").show();
        $("#fieldOneLabel").text("Amount");
        $("#fieldTwoLabel").text("Stock Symbol");
        break;
      case "CANCEL_SET_BUY":
        $("#fieldOne").hide();
        $("#fieldTwo").show();
        $("#fieldTwoLabel").text("Stock Symbol");
        break;
      case "SET_BUY_TRIGGER":
        $("#fieldOne").show();
        $("#fieldTwo").show();
        $("#fieldOneLabel").text("Amount");
        $("#fieldTwoLabel").text("Stock Symbol");
        break;
      case "SET_SELL_AMOUNT":
        $("#fieldOne").show();
        $("#fieldTwo").show();
        $("#fieldOneLabel").text("Amount");
        $("#fieldTwoLabel").text("Stock Symbol");
        break;
      case "SET_SELL_TRIGGER":
        $("#fieldOne").show();
        $("#fieldTwo").show();
        $("#fieldOneLabel").text("Amount");
        $("#fieldTwoLabel").text("Stock Symbol");
        break;
      case "CANCEL_SET_SELL":
        $("#fieldOne").hide();
        $("#fieldTwo").show();
        $("#fieldTwoLabel").text("Stock Symbol");
        break;
      case "DUMPLOG":
        $("#fieldOne").hide();
        $("#fieldTwo").show();
        $("#fieldTwoLabel").text("Filename");
        break;
      case "DISPLAY_SUMMARY":
        $("#fieldOne").hide();
        $("#fieldTwo").hide();
        break;
    }
  });

  $("#commandForm").submit(function(event) {
    console.log("this is ", this);
    console.log("submit action captured");
    event.preventDefault(); //prevent default action
    var post_url = $(this).attr("action"); //get form action url
    var form_data = $(this).serialize(); //Encode form elements for submission
    console.log("datafield 1 is ", $("input[name=numberInput]").val());
    console.log("datafield 2 is ", $("input[name=stringInput]").val());
    $.post(post_url, form_data, function(response) {
      $("#server-results").html(response);
    });
    submitRequest();
  });

  // $("#submitBtn").on("click", submitRequest);
  // $("#stringInput").on("click", () => $("#stringInput").val(""));
  // $("#numberInput").on("click", () => $("#numberInput").val(""));
});

function submitRequest() {
  currentCommand = $(".btn-group > .btn.active").text();
  console.log("in submitRequest, id is " + currentCommand);
  stringInput = $("input[name=stringInput]").val();
  numberInput = $("input[name=numberInput]").val();
  console.log("string input is ", stringInput, "number input is ", numberInput);

  $.ajax({
    type: "POST",
    url: "sendCommand",
    data: {
      command: currentCommand,
      username: username,
      amount: numberInput,
      string: stringInput
    },
    cache: false,
    success: function(results) {
      console.log(results);
      $('#displayDiv p').text("result: "+results);
      $('#errorDiv p').text("error: none");
    },
    error: function(jqXHR, textStatus, errorThrown) {
      $('#result').html('<p>status code: '+jqXHR.status+'</p><p>errorThrown: ' + errorThrown + '</p><p>jqXHR.responseText:</p><div>'+jqXHR.responseText + '</div>');
      console.log('jqXHR:');
      console.log(jqXHR);
      console.log('textStatus:');
      console.log(textStatus);
      console.log('errorThrown:');
      console.log(errorThrown);

      $('#displayDiv p').text("result: None");
      $('#errorDiv').html('<p>status code: '+jqXHR.status+'</p><p>errorThrown: ' + errorThrown + '</p><p>jqXHR.responseText:</p><div>'+jqXHR.responseText + '</div>');
      
  },
  })
    // .done(function() {
    //   alert("success");
    // })
    // .fail(function() {
    //   alert("error");
    // });
}