
var UpdateTimerId = 0;

function startTask() {
    $.ajax({
	url: '/api/start',
	type: 'GET',
	dataType: 'json',
	cache: false,
    }).done(function (res) {
	if (res['result']) {
	    syncStatus();
	} else {
	    alart('start task error');
	}
    });
}

function updateStatus(dataSet) {
    var tbody = $('table tbody');
    // clear tbody
    tbody.find('tr').remove();

    var taskProgress = false;
    for (var i = 0; i < dataSet.length; i++) {
	var data = dataSet[i];
	var trtag = "<tr>";
	if (!data['done']) {
	    trtag = "<tr class=\"active\">";
	}
	var par = data['progress']*10;
	var progress = '<div class="progress-bar" role="progressbar" area-valuenow="'+par+'" area-valuemin="0" area-valuemax="100" style="width: '+par+'%;">'+par+'%</div>'
	var html = trtag+"<td>"+data['id']
	    +"</td><td><div class=\"progress\">"+progress
	    +"</div></td><td>"+data['done']
	    +"</td></tr>";
	tbody.append(html)
	if (!data['done']) {
	    taskProgress = true;
	}
    }
    if (taskProgress) {
	UpdateTimerId = setTimeout("syncStatus()", 1000);
    }
}

function syncStatus() {
    clearTimeout(UpdateTimerId);
    $.ajax({
	url: '/api/status',
	type: 'GET',
	dataType: 'json',
	cache: false,
    }).done(function (res) {
	updateStatus(res);
    });
}
