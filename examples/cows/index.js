var cows = require('cows')
var all = cows();

function show(cow) {
    var cowText = all[cow];
    console.log(cowText);
}

var cowArgument = process.argv[2]
if(!cowArgument || isNaN(Number(cowArgument))) {
  var randomCow = Math.floor((Math.random() * all.length) + 1);
  show(randomCow);
}
else {
  var cowNumber = Number(cowArgument);
  if(cowNumber && cowNumber <= all.length) {
    show(cowNumber);
  } else {
    console.log("Enter cow number between [1-"+all.length+"]")
  }
}
