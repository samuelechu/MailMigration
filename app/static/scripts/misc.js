$(document).ready(function(){
	$('[data-toggle="tooltip"]').tooltip();   
});

function jobInProgress(uid, callback) {
	var xhr = new XMLHttpRequest();
	xhr.open('GET', 'jobInfo?uid=' + uid);
	xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
	xhr.onload = function() {
		console.log('Job in progress : ' + xhr.responseText);
		var resp = JSON.parse(xhr.responseText);
		callback(uid, resp.Source_id != '')
	};
	xhr.send();

}

function manageSections(uid, jobInProgress) {
	if(!jobInProgress){
		$("#selectSection").collapse('show');

		var sourceName = document.getElementById("sourceName").innerHTML
		var destName = document.getElementById("destName").innerHTML

		if( sourceName != "" && destName != "") {
			$("#transferButtonSection").collapse('show');
		} else {
			$("#transferButtonSection").collapse('hide');
		}

	} else {
		if (localStorage.getItem('visited') == 'true') { 
   			$("#cancelJob").hide();
		}
		$("#markFails").hide();
		$("#jobSection").collapse('show');
		if (window.Worker){
			var progressUpdater = new Worker("scripts/progressUpdater.js")


			progressUpdater.onmessage = function(e) {
				console.log(e.data.percentage)


				$("#jobProgressBar").css('width', e.data.percentage + '%');
				$('#jobProgressBar').html(Math.floor(e.data.percentage) + '%');

				if (e.data.percentage > 0) {
					$('#transferProgress').html("Email threads processed: " + e.data.processed + "/" + e.data.total + ", Threads failed to transfer: " + e.data.failed + "<sup><span id=\"failedInfo\" class=\"glyphicon glyphicon-question-sign\" data-toggle=\"tooltip\" data-placement=\"left\" title=\"Emails may fail to transfer when they are too large. If this happens, an option to mark these emails will be given after processing is finished.\"></span></sup>");
				}  

				if (e.data.percentage == 100){
					$("#cancelJob").hide();
					if (localStorage.getItem('visited') == 'false' && e.data.failed != 0) { 
	           			$("#markFails").show();
	        		}
					$('#finishSection').collapse('show');
				}

			};

			var uidMessage = { uid: uid };
			progressUpdater.postMessage(uidMessage)

		}
	}
}
