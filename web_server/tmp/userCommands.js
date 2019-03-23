$(document).ready(function() {
  $("#myButtons :input").change(function() {
    console.log("button clicked");
    console.log(this.id); // points to the clicked input button

    switch (this.id) {
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
});
