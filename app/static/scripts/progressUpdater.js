//webworker

self.addEventListener("message", function(e) {
	updateProgress(e.data.uid)
    // the passed-in data is available via e.data
}, false);

function updateProgress(uid){

	var xhr = new XMLHttpRequest();
	xhr.open('GET', '../../jobInfo?uid=' + uid);
	xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
  	xhr.onload = function() {
    console.log('Job in progress : ' + xhr.responseText);
    var resp = JSON.parse(xhr.responseText);
   
    var percentage = resp.Processed_threads * 100 / resp.Total_threads
    if(resp.Processed_threads == 0 && resp.Total_threads == 0){
    	percentage = 0
    }
    
    console.log('percentage: ' + percentage)
    var percentageMessage = {percentage: percentage, processed: resp.Processed_threads, total: resp.Total_threads, failed: resp.Failed_threads};

    postMessage(percentageMessage)
    if(percentage < 100){
		setTimeout(function() {
			updateProgress(uid)
		}, 5000);
	}
  };
  xhr.send();
  
}