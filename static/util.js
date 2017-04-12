/*
Copyright 2017 Favyen Bastani <fbastani@perennate.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

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
