function ButtonAPICall(elem, url, data, success, sure) {
	$(elem).click(event, function() {
		if (sure) {
			if (!confirm(sure)) {
				return;
			}
		}
		
		$.ajax({
			url: url,
			dataType: "json",
			data: data,
			success: success,
			error: function(jqXHR) {
				alert("Error: " + jqXHR.responseText);
			}
		})
	});
}

var AjaxErrorFunc = function(jqXHR) {
	alert("Error: " + jqXHR.responseText);
};

$(document).ready(function(){
	
//	// Connect all sensors button
//	ButtonAPICall($("button[data-role=connect-all-sensors]"), "/api/set/sensors/connect-all", {}, function(){
//		location.reload();
//	});
//	
//	// Disconnect all sensors button
//	ButtonAPICall($("button[data-role=disconnect-all-sensors]"), "/api/set/sensors/disconnect-all", {}, function(){
//		location.reload();
//	});
	
	// Connect one sensor
	$("button[data-role=connect-sensor]").each(function(){
		ButtonAPICall(this, "/api/set/sensors/connect", { sensor: $(this).attr("data-sensor") }, function() {
			location.reload();
		});
	});
	
	// Disconnect one sensor
	$("button[data-role=disconnect-sensor]").each(function() {
		ButtonAPICall(this, "/api/set/sensors/disconnect", { sensor: $(this).attr("data-sensor") }, function() {
			location.reload();
		});
	});
	
	// Start one sensor
	$("button[data-role=start-sensor]").each(function() {
		ButtonAPICall(this, "/api/set/sensors/start", { sensor: $(this).attr("data-sensor") }, function() {
			location.reload();
		});
	});
	
	// Stop one sensor
	$("button[data-role=stop-sensor]").each(function() {
		ButtonAPICall(this, "/api/set/sensors/stop", { sensor: $(this).attr("data-sensor") }, function() {
			location.reload();
		});
	});
	
	// Run log in real time
	$("button[data-role=start-logread-realtime]").each(function() {
		var el = $(this);
		var logName = $(this).parent().parent().attr("data-logname");
		ButtonAPICall(this, "/api/set/sensorlogs/start-logread-realtime", { logname: logName }, function() {
			location.reload();
		}, "Are you sure you want to start a log in real time?");
	});
	
	// Stop log in real time
	$("button[data-role=stop-logread-realtime]").each(function() {
		ButtonAPICall(this, "/api/set/sensorlogs/stop-logread-realtime", {}, function() {
			location.reload();
		}, "Are you sure you want to stop the log reading?");
	})
	
	// Delete sensor log
	$("[data-role=delete-sensorlog]").each(function() {
		var el = $(this);
		var logName = $(this).parent().parent().attr("data-logname");
		ButtonAPICall(this, "/api/set/sensorlogs/delete", { logname: logName }, function() {
			el.parent().parent().hide("slow");
		}, "Are you sure you want to delete " + logName + "?");
	});
	
	// Rename sensor log
	$("[data-role=rename-sensorlog]").each(function() {
		var el = $(this);
		var logName = $(this).parent().parent().attr("data-logname");
		el.click(function() {
			var newName = prompt("New name:");
			if (newName == "" || !newName) {
				return;
			}
			$.ajax({
				url: "/api/set/sensorlogs/rename",
				dataType: "json",
				data: {logname: logName, newname: newName},
				error: function(jqXHR) { alert("Error: " + jqXHR.responseText); },
				success: function(data) {
					el.parent().parent().find(".name").html(newName + ".log");
				}
			})
		});
	});
	
	// Stop SLAM buttons
	$("button[data-role=stop-slam]").click(function(){
		if (!confirm("Are you sure you want to stop SLAM?")) {
			return;
		}
		
		$.ajax("/api/set/slam/stop/", {
			error: function(jqXHR, textStatus, errorThrown) {
				alert("Error: " + textStatus + errorThrown);
			},
			success: function(data) {
				location.reload();
			}
		});
	});
	
	// Start SLAM buttons
	$("button[data-role=start-slam]").click(function() {
		$.ajax("/api/set/slam/start/", {
			error: function(jqXHR, textStatus, errorThrown) {
				alert("Error: " + textStatus + errorThrown);
			},
			success: function(data) {
				location.reload();
			}
		});
	});
	
	// // Terminate SLAM buttons
	// $("button[data-role=terminate-slam]").click(function() {
	// 	if (!confirm("Are you sure you want to terminate SLAM?")) {
	// 		return;
	// 	}
		
	// 	$.ajax("/api/set/slam/terminate/", {
	// 		error: function(jqXHR, textStatus, errorThrown) {
	// 			alert("Error: " + textStatus + errorThrown);
	// 		},
	// 		success: function(data) {
	// 			location.reload();
	// 		}
	// 	});
	// });
	
	// Start all sensors
	$("button[data-role=start-all-sensors]").click(function() {
		$.ajax("/api/set/sensors/start-all/", {
			error: function(jqXHR, textStatus, errorThrown) {
				alert("Error: " + textStatus + errorThrown);
			},
			success: function(data) {
				location.reload();
			}
		});
	});
	
	// Stop all sensors
	$("button[data-role=stop-all-sensors]").click(function() {
		$.ajax("/api/set/sensors/stop-all/", {
			error: function(jqXHR, textStatus, errorThrown) {
				alert("Error: " + textStatus + errorThrown);
			},
			success: function(data) {
				location.reload();
			}
		});
	});
});