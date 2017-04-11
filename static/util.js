/**
 * Shuffles array in place.
 * @param {Array} a items The array containing the items.
 * Source: http://stackoverflow.com/questions/6274339/how-can-i-shuffle-an-array
 */
function shuffle(a) {
    var j, x, i;
    for (i = a.length; i; i--) {
        j = Math.floor(Math.random() * i);
        x = a[i - 1];
        a[i - 1] = a[j];
        a[j] = x;
    }
}

function time() {
	return new Date().getTime();
}

function getMean(a) {
	var s = 0;
	a.forEach(function(x) {
		s += x;
	});
	return s / a.length;
}

function getStddev(a) {
	var mean = getMean(a);
	var s = 0;
	a.forEach(function(x) {
		var d = x - mean;
		s += d * d;
	});
	return Math.sqrt(s / a.length);
}
