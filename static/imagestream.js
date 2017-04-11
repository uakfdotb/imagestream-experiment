var QUALIFICATION_TASK_LENGTH = 200;
var QUALIFICATION_POSITIVE_IMAGES = 25;
var QUALIFICATION_MAX_CLICK_DELAY = 800;
var QUALIFICATION_START_DELAY = 10;
var QUALIFICATION_REQUIRED_PRECISION = 0.9;
var QUALIFICATION_REQUIRED_RECALL = 0.6;
var RAPID_IMAGE_DELAY = 200;

function getQualificationTask() {
	var images = [];
	for(var i = 0; i < QUALIFICATION_TASK_LENGTH; i++) {
		images.push({
			'id': i,
			'image_dir': 'qual',
			'image_id': i,
		});
	}
	shuffle(images);
	return {
		'type': 'rapid',
		'images': images,
	};
}

function showStage(name) {
	$('.stageDiv').css('display', 'none');
	$('#qualInstructions').css('color', 'green');
	$('#' + name).css('display', '');
}

function preloadTask(task, callback) {
	// preloaded images will map from display index to an Image object
	var preloadedImages = {};
	
	var finished = 0;
	var total = 0;
	var preloadOne = function(idx, path) {
		var image = new Image();
		total++;
		image.onload = function() {
			finished++;
			if(finished == total) {
				callback(preloadedImages);
			}
		};
		image.onerror = function() {
			alert('error loading from ' + path);
		};
		image.src = path;
		preloadedImages[idx] = image;
	};
	for(var i = 1; i <= QUALIFICATION_START_DELAY; i++) {
		preloadOne(-i, 'images/counter/' + i + '.png');
	}
	task.images.forEach(function(image, i) {
		preloadOne(i, 'images/' + image.image_dir + '/' + image.image_id + '.png');
	});
}

function runTask(task, preloadedImages, callback) {
	showStage(task.type + 'Task');
	if(task.type == 'rapid') {
		runRapidTask(task, preloadedImages, callback);
	} else if(task.type == 'conventional') {
		runConventionalTask(task, preloadedImages, callback);
	}
}

function runRapidTask(task, preloadedImages, callback) {
	var counter = -QUALIFICATION_START_DELAY;
	var startTime = time();
	var clickLog = [];
	var displayLog = [];
	var updateIntervalObj;
	
	// create handler to listen for spacebar presses
	// on spacebar, we want to update clickLog and also display the last four images under rapidSelectedImageDiv
	var pressHandler = function(e) {
		if(e.which != 32) {
			return;
		}
		clickLog.push({'time': time() - startTime});
		
		$('#rapidSelectedImageDiv').children().remove();
		for(var i = Math.max(0, displayLog.length - 4); i < displayLog.length; i++) {
			var img = $('<img>').attr('src', preloadedImages[i].src)
			                    .attr('width', '100px')
			                    .attr('height', '100px');
			img.appendTo($('#rapidSelectedImageDiv'));
		}
		
		// clear after one second, but only if the user hasn't clicked again
		var clickID = clickLog.length;
		setTimeout(function() {
			if(clickID == clickLog.length) {
				$('#rapidSelectedImageDiv').children().remove();
			}
		}, 1000);
	};
	$('body').on('keydown', pressHandler);
	
	// finish function will be executed when we are done running through images
	// here, we cleanup and then callback
	var finish = function() {
		$('body').off('keydown', pressHandler);
		showStage('finishTask');
		if(updateIntervalObj) {
			window.clearInterval(updateIntervalObj);
		}
		$('#rapidMainImage').attr('src', 'images/blank.png');
		setTimeout(function() {
			callback(time() - startTime, displayLog, clickLog);
		}, 1000);
	};
	
	// function for timer to display the next image in sequence
	var showNextImage = function() {
		if(counter >= task.images.length) {
			// we're done!
			finish();
			return;
		}
		$('#rapidMainImage').attr('src', preloadedImages[counter].src);
		if(counter >= 0) {
			displayLog.push({
				'image_id': task.images[counter].image_id,
				'time': time() - startTime,
			});
		}
		counter++;
	};

	updateIntervalObj = setInterval(showNextImage, RAPID_IMAGE_DELAY);
}

function runConventionalTask(task, preloadedImages, callback) {
	var startTime = time();
	var imageIndex = 0;
	var labels = [];
	var finish;
	
	var showImage = function() {
		$('#conventionalMainImage').attr('src', preloadedImages[imageIndex].src);
	};
	
	// setLabel is called whenever the user inputs a label
	var setLabel = function(label) {
		labels.push({
			'task_image_id': task.images[imageIndex].id,
			'is_in_class': label,
		});
		imageIndex++;
		if(imageIndex >= task.images.length) {
			// we're done!
			finish();
			return;
		}
		showImage();
		console.log(labels);
	};
	
	// create handlers to listen for button/key presses
	var buttonHandler = function(e) {
		console.log($(this).data());
		if($(this).data('label')) {
			setLabel(true);
		} else {
			setLabel(false);
		}
	};
	var pressHandler = function(e) {
		if(String.fromCharCode(e.which).toUpperCase() == 'D') {
			setLabel(true);
		} else if(String.fromCharCode(e.which).toUpperCase() == 'O') {
			setLabel(false);
		}
		console.log(String.fromCharCode(e.which));
	};
	$('.conventionalBtn').on('click', buttonHandler);
	$('body').on('keydown', pressHandler);
	
	// finish function will be executed when we are done running through images
	// here, we cleanup and then callback
	var finish = function() {
		$('.conventionalBtn').off('click', buttonHandler);
		$('body').off('keydown', pressHandler);
		showStage('finishTask');
		$('#rapidMainImage').attr('src', 'images/blank.png');
		setTimeout(function() {
			callback(time() - startTime, labels);
		}, 1000);
	};
	
	showImage();
}

$(document).ready(function() {
	var tasks = [];
	var nextTaskPreloadedImages = null;
	var qualMean = null;
	var qualSigma = null;
	
	var nextTask = function() {
		if(tasks.length > 0) {
			showStage(tasks[0].type + 'Instructions');
			$('.taskStart').prop('disabled', true);
			$('.taskStart').text('Start (loading...)');
			preloadTask(tasks[0], function(preloadedImages) {
				nextTaskPreloadedImages = preloadedImages;
				$('.taskStart').prop('disabled', false);
				$('.taskStart').text('Start');
			});
		} else {
			showStage('endInstructions');
		}
	};
	
	var fetchTasks = function() {
		$.post('/start-experiment', function(data) {
			data.forEach(function(task) {
				task.images.forEach(function(image) {
					image.image_dir = 'data';
				});
				tasks.push(task);
			});
			console.log('got ' + tasks.length + ' tasks from server');
			nextTask();
		}, 'json');
	};
	
	// preload qualification task
	$('.qualStart').prop('disabled', true);
	$('.qualStart').text('Start (loading...)');
	var qualTask = getQualificationTask();
	preloadTask(qualTask, function(preloadedImages) {
		nextTaskPreloadedImages = preloadedImages;
		$('.qualStart').prop('disabled', false);
		$('.qualStart').text('Start');
	});
	
	$('.qualStart').click(function() {
		runTask(qualTask, nextTaskPreloadedImages, function(duration, displayLog, clickLog) {
			// compute precision and recall
			// we consider a click to be matched if it is within QUALIFICATION_MAX_CLICK_DELAY of a positive example
			var numActual = QUALIFICATION_POSITIVE_IMAGES;
			var numPredict = clickLog.length;
			var numMatch = 0;
			var delaySamples = [];
			clickLog.forEach(function(click) {
				var matchedWith = null;
				displayLog.forEach(function(display) {
					if(display.image_id >= QUALIFICATION_POSITIVE_IMAGES) {
						// this is a negative image
						return;
					} else if(click.time <= display.time || click.time >= display.time + QUALIFICATION_MAX_CLICK_DELAY) {
						// clicked before this image or too long after this image
						return;
					}
					if(!matchedWith || display.time > matchedWith.time) {
						matchedWith = display;
					}
				});
				if(matchedWith) {
					numMatch++;
					delaySamples.push(click.time - matchedWith.time);
				}
			});
		
			// compute mean/sigma from delaySamples
			qualMean = getMean(delaySamples);
			qualSigma = getStddev(delaySamples);

			var precision = numMatch / numPredict;
			var recall = numMatch / numActual;
			console.log('precision=' + precision + ', recall=' + recall + ', mean=' + qualMean + ', sigma=' + qualSigma);
			console.log(clickLog);
			console.log(displayLog);
			if(precision < QUALIFICATION_REQUIRED_PRECISION || recall < QUALIFICATION_REQUIRED_RECALL) {
				// failed qualification
				showStage('endInstructions');
				return;
			}
		
			// grab tasks from server and begin
			fetchTasks();
		});
	});
	
	$('.taskStart').click(function() {
		var task = tasks.shift();
		if(task.type == 'rapid') {
			runTask(task, nextTaskPreloadedImages, function(duration, displayLog, clickLog) {
				var data = {
					'task_id': task.id,
					'worker_id': task.worker_id,
					'duration': duration,
					'rapid_logs': [{
						'display_json': JSON.stringify(displayLog),
						'click_json': JSON.stringify(clickLog),
						'mean': qualMean,
						'sigma': qualSigma,
					}],
				};
				$.post('/end-experiment', JSON.stringify(data), function() {
					nextTask();
				});
			});
		} else if(task.type == 'conventional') {
			runTask(task, nextTaskPreloadedImages, function(duration, labels) {
				var data = {
					'task_id': task.id,
					'worker_id': task.worker_id,
					'duration': duration,
					'conventional_labels': labels,
				};
				$.post('/end-experiment', JSON.stringify(data), function() {
					nextTask();
				});
			});
		}
	});
});
