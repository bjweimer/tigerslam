// Slam is responsible for communicating with the server and updating the
// interface. Some elements are assumed to be present. It is implemented as a
// jQuery plugin, which takes a map from the Google Maps Javascript API v3 as
// input.
;(function($) {
	
	$.slamstats = function(map, options) {
		
		// Default settings values
		var defaults = {
			url: "/api/get/slam/stats",
			pullRate: 1000,
			mapUpdateDistanceDiff: 5.0,
			mapOffset: {
				X: 0,
				Y: 0,
			},
			robot: {
				physical: [
					[
						{X: -5, Y: -2.5}, {X: 5, Y: -2.5},
						{X: 5, Y: 2.5}, {X: -5, Y: 2.5}
					],
					[
						{X: 2.5, Y: -2.5}, {X: 5, Y: 0}, {X: 2.5, Y: 2.5}
					]
				],
				fillColor: "#0000FF",
				fillOpacity: 0.10,
				strokeColor: "#0000FF",
				strokeOpacity: 0.8,
				strokeWeight: 0.2,
				toggler: false,	
			},
			trace: {
				line: {
					strokeColor: '#f00',
					strokeOpacity: 1.0,
					strokeWeight: 1,
				},
				toggler: false,
			},
			path: {
				line: {
					strokeColor: '#0f0',
					strokeOpacity: 1.0,
					strokeWeight: 1,
				}
			},
			follow: {
				toggler: false,
			}
		};
		
		var plugin = this;
		
		plugin.settings = {};
		plugin.contextMenu = {
			'latLng': false,
			'physical': false,
			'point': false,
		};
		
		// Initialize the plugin -- This is done when the plugin is loaded. 
		var init = function() {
			plugin.settings = $.extend({}, defaults, options);
			plugin.map = map;
			plugin.ruler = false;
			plugin.closeModal = $("#close-modal").modal({show: false});
			plugin.progressModal = new $.progressmodal($("#progress-modal"));

			// Set up trace toggling
			$(plugin.settings.trace.toggler).change(function() {
				if ($(this).is(":checked")) {
					plugin.trace.setMap(plugin.map);
				} else {
					plugin.trace.setMap(null);
				}
			});
			
			// Set up robot polygon toggling
			$(plugin.settings.robot.toggler).change(function() {
				if ($(this).is(":checked")) {
					plugin.robot.setMap(plugin.map);
				} else {
					plugin.robot.setMap(null);
				}
			});
			
			// Set up map context menu (right click)
			plugin.contextMenu = google.maps.event.addListener(plugin.map, "rightclick", function(event) {
				updateContextMenu(event);
				$("#context_menu").fadeIn({duration: 100});
			});

			// Set up ruler
			plugin.ruler = new google.maps.Polyline({
				map: plugin.map,
				strokeColor: "#00FF00",
				strokeOpacity: 1.0,
				strokeWeight: 3.0,
				editable: true,
				draggable: true,
				geodesic: false,
			});
			plugin.ruler.infoWindow = new google.maps.InfoWindow();
			google.maps.event.addListener(plugin.ruler.getPath(), "insert_at", updateRuler);
			google.maps.event.addListener(plugin.ruler.getPath(), "remove_at", updateRuler);
			google.maps.event.addListener(plugin.ruler.getPath(), "set_at", updateRuler);

			// Path
			plugin.path = new google.maps.Polyline(plugin.settings.path.line);
			plugin.path.setMap(plugin.map);

			// Context menu functions: a map between the roles of the links in
			// the context menu and the function they trigger.
			var contextMenuFunctions = {
				// Center the map at this coordinate
				'center': function() {
					plugin.map.panTo(plugin.contextMenu.latLng);
				},
				// Zoom in and center at this coordinate
				'zoom-in': function() {
					plugin.map.setZoom(plugin.map.getZoom() + 1);
					plugin.map.setCenter(plugin.contextMenu.latLng);
				},
				// Zoom out and center at this coordinate
				'zoom-out': function() {
					plugin.map.setZoom(plugin.map.getZoom() - 1);
					plugin.map.setCenter(plugin.contextMenu.latLng);
				},
				// Set this point as a goal for the robot to navigate to
				'go-here': function() {
					goTo(plugin.contextMenu.physical);
				},
				// Update the map
				'update-map': updateMap,
				// Insert ruler-from
				'insert-ruler-point': function() {
					plugin.ruler.getPath().push(plugin.contextMenu.latLng);
				},
				'reset-ruler': function() {
					plugin.ruler.getPath().clear();
				}
			};

			for (var role in contextMenuFunctions) {
				$("#context_menu").find("[role=" + role + "]").click(contextMenuFunctions[role]);
			}
			
			$("body").click(function(){
				$("#context_menu").fadeOut({duration: 100});
			});
			
			// Set up pathplanning modal, which should display whenever the user
			// is initializing a path planning process on the server.
			$("#pathplanning-modal").modal({
				keyboard: false,
				backdrop: "static",
				show: false,
			});
			
			// Set up trace
			plugin.trace = new google.maps.Polyline(plugin.settings.trace.line);
			plugin.settings.trace.toggler.change();

			// Set up a robot polygon, draw the trace
			plugin.robot = new google.maps.Polygon(plugin.settings.robot);
			plugin.settings.robot.toggler.change();

			// Close button
			$("[data-role=close-slam]").click(showCloseModal);
			$("[data-role=close-without-saving]").click(terminate);
			$("[data-role=save-and-close]").click(saveAndClose);
		};
		
		// Start the continuous updation of the interface.
		plugin.Start = function() {
			// Set up update interval
			plugin.interval = setInterval(update, plugin.settings.pullRate);
		}
		
		// Call the api and update all data in the DOM
		var update = function() {
			$.getJSON(plugin.settings.url, function(data) {
				// Update sidebar data
				$("[data-role=position-x]").html(twodecimals(data.position.X));
				$("[data-role=position-y]").html(twodecimals(data.position.Y));
				$("[data-role=position-theta]").html(twodecimals(rad2deg(data.position.Theta)));
				
				// Update trace
				if (plugin.trace) {
					updateTrace(data.position);
				}
				
				// Decide if we should update the map, then do it
				if (!plugin.lastMapUpdatePosition) {
					plugin.lastMapUpdatePosition = data.position;
				} else {
					if (positionDistance(data.position, plugin.lastMapUpdatePosition) > plugin.settings.mapUpdateDistanceDiff) {
						updateMap();
						plugin.lastMapUpdatePosition = data.position;
					}
				}

				// Update the robot
				if (plugin.robot && plugin.robot.getMap()) {
					updateRobot(data.position);
				}

				// Robot following
				if (plugin.settings.follow.toggler) {
					if (plugin.settings.follow.toggler.is(":checked")) {
						centerAt(data.position);
					}
				}

				// Show appropriate buttons
				if (data.state == "STOPPED") {
					$("button[data-role=stop-slam]").hide();
					$("button[data-role=close-slam]").show();
					$("button[data-role=start-slam]").show();
				} else if (data.state = "RUNNING") {
					$("button[data-role=stop-slam]").show();
					$("button[data-role=close-slam]").hide();
					$("button[data-role=start-slam]").hide();
				}
			});
		};

		// Show the close-modal
		var showCloseModal = function() {
			// Show the close modal
			plugin.closeModal.modal('show');
		};

		// Save and close. While the map is saved, display a progress bar.
		// When the map is saved, terminate the slam process and reload.
		var saveAndClose = function() {

			plugin.closeModal.modal('hide');
			plugin.progressModal.Run("Saving map ...", 30 * 1000);

			saveMap(function() {
				// When saved, terminate
				plugin.progressModal.Finished(terminate);
			});

		}

		// Gather meta information and request that the server saves the map.
		var saveMap = function(callback) {

			var name = $("#map-name").val();
			var description = $("#map-description").val();

			// Send ajax query
			$.ajax("/api/set/slam/save", {
				data: {
					name: name,
					description: description,
				},
				success: callback,
				error: function(jqXHR) { alert("Error: " + jqXHR.responseText); },
			});

		}

		// Close without saving
		var terminate = function() {
			$.ajax("/api/set/slam/terminate", {
				error: function(jqXHR) { alert("Error: " + jqXHR.responseText); },
				success: function(data) { location.reload(); },
			});
		}

		// Update the car marker in the map from a position given in physical
		// coordinates.
		var updateCarMarker = function(position) {
			plugin.carMarker.setPosition(
					plugin.map.getProjection().fromPhysicalToLatLng(
							plugin.settings.mapOffset.X + position.X,
							plugin.settings.mapOffset.Y + position.Y
							)
					);
		};
		
		// Update the trace with a new point (results in a new line segment).
		var updateTrace = function(position) {
			plugin.trace.getPath().push(
					plugin.map.getProjection().fromPhysicalToLatLng(
							plugin.settings.mapOffset.X + position.X,
							plugin.settings.mapOffset.Y + position.Y
							)
					);
		};
		
		// Trigger a map update -- flushes the old map tiles and replaces them
		// with new ones.
		var updateMap = function() {
			plugin.map.mapTypes.slam.update();
		};
		
		// The distance between two coordinates
		var positionDistance = function(pos1, pos2) {
			return Math.sqrt(Math.pow(pos1.X - pos2.X, 2) + Math.pow(pos1.Y - pos2.Y, 2));
		};
		
		// Update the context menu; Set the displayed coordinates in the menu
		// and give it an appropriate position.
		var updateContextMenu = function(event) {
			var menu = $("#context_menu");
			var point = plugin.map.getProjection().fromLatLngToPoint(event.latLng);
			var physical = plugin.map.getProjection().fromPointToPhysical(point);
			
			plugin.contextMenu.latLng = event.latLng;
			plugin.contextMenu.point = point;
			plugin.contextMenu.physical = {
					x: physical.x - plugin.settings.mapOffset.X,
					y: physical.y - plugin.settings.mapOffset.Y,
			};
			
			// Set up position
			menu.css("left", event.pixel.x);
			menu.css("top", event.pixel.y + 60);
			
			menu.find("[role=coordinates]").html(
					"x: " + twodecimals(plugin.contextMenu.physical.x) + 
					" y: " + twodecimals(plugin.contextMenu.physical.y));
		};
		
		// Go to
		// Send a command to the server, setting this point as a goal position
		// which the robot should navigate to. While the path is planned,
		// display a progress bar, as it may take some time.
		var goTo = function(coordinates) {
			
			var modal = $("#pathplanning-modal");
			var progressBar = modal.find(".progress .bar");
			
			// A function to update the progress
			function updateProgress() {
				var progress = ((new Date()).getTime() - startTime.getTime()) / duration * 100;
				progressBar.css("width", "" + progress + "%");
				
				// We're at 100 %, close modal
				if (progress >= 100) {
					clearInterval(interval);
					modal.modal("hide");
				}
			}
			
			// What to do when the api call returns
			function success() {
				clearInterval(interval);
				modal.modal("hide");
			}
			
			// For now, assume it takes constant time duration in sec.
			var duration = 5000;
			var startTime = new Date();
			
			modal.modal('show');
			var interval = setInterval(updateProgress, 100);
			
			// Actually send the request
			$.ajax({
				url: "/api/set/motor/goto",
				dataType: "json",
				data: coordinates,
				success: function(data) {
					success();
					updatePath();
				},
				error: function(jqXHR) {
					clearInterval(interval);
					modal.modal("hide");
					alert("Error: " + jqXHR.responseText);
				}
			})
		};

		// Get the current path from the server, draw it.
		var updatePath = function() {
			
			// Get path
			$.ajax("/api/get/motor/path", {
				dataType: "json",
				success: function(data) {
					drawPath(data);
				},
				error: function(jqXHR) {
					alert("Error: " + jqXHR.responseText);
				}
			});

		};

		// Draw a path which the robot should traverse. The path is assumed to
		// be an array of coordinates specified as {x: number, y: number}
		// objects.
		var drawPath = function(path) {
			plugin.path.getPath().clear()
			
			var projection = plugin.map.getProjection();

			for (var i in path.Poses) {
				var latLng = projection.fromPhysicalToLatLng(
					plugin.settings.mapOffset.X + path.Poses[i][0],
					plugin.settings.mapOffset.Y + path.Poses[i][1]
					);
				plugin.path.getPath().push(latLng);
			}

		};

		// Center map at given coordinates
		var centerAt = function(pose) {
			var latLng = plugin.map.getProjection().fromPhysicalToLatLng(pose.X + plugin.settings.mapOffset.X, pose.Y + plugin.settings.mapOffset.Y);
			plugin.map.panTo(latLng);
		};

		// Draw a robot as a polygon at a given position and rotation. Pose is
		// given as {x: number, y: number, theta: number}.
		var updateRobot = function(pose) {
			var projection = plugin.map.getProjection();
			var physical = plugin.settings.robot.physical;

			var paths = [];
			for (var i in physical) {
				paths[i] = [];
				for (var j in physical[i]) {
					transformed = transformedPoint(physical[i][j], pose);
					paths[i][j] = projection.fromPhysicalToLatLng(transformed.X + plugin.settings.mapOffset.X, transformed.Y + plugin.settings.mapOffset.Y);
				}
			}

			// Remove robot from the map, redefine it's path and replace it in the map
			plugin.robot.setMap(null);
			plugin.robot.setPaths(paths);
			plugin.robot.setMap(plugin.map);
		};

		// Calculate the total distance of the ruler (which is a polyline)
		var updateRuler = function() {
			var path = plugin.ruler.getPath();
			var distance = getPolylineDistance(path);

			if (distance == 0) {
				return;
			}

			var lastPoint = path.getAt(path.getLength() - 1);
			plugin.ruler.infoWindow.setContent('<strong>Distance:</strong> ' + twodecimals(distance) + ' meters');
			plugin.ruler.infoWindow.setPosition(lastPoint);

			plugin.ruler.infoWindow.open(plugin.map);
		}

		// Get the total distance of a polyline
		var getPolylineDistance = function(path) {
			var projection = plugin.map.getProjection();

			if (path.getLength() < 2) {
				return 0;
			}

			var distance = 0;
			for (var i = 0; i < path.getLength() - 1; i++) {
				var a = projection.fromPointToPhysical(projection.fromLatLngToPoint(path.getAt(i)));
				var b = projection.fromPointToPhysical(projection.fromLatLngToPoint(path.getAt(i + 1)));
				distance += Math.sqrt(Math.pow(a.x - b.x, 2) + Math.pow(a.y - b.y, 2));
			}
			return distance;
		}

		// Transform a point (translate, rotate) relative to some pose.
		var transformedPoint = function(point, pose) {
			return {
				X: point.X * Math.cos(pose.Theta) - point.Y * Math.sin(pose.Theta) + pose.X,
				Y: point.X * Math.sin(pose.Theta) + point.Y * Math.cos(pose.Theta) + pose.Y,
				Theta: pose.Theta + point.Theta,
			};
		}

		// Return the number rounded to two decimals.
		var twodecimals = function(number) {
			return Math.round(number * 100) / 100;
		};
		
		// Convert radians to degrees and normalize to [-180, 180]
		var rad2deg = function(rad) {
			rad = rad % (2 * Math.PI);
			if (rad > Math.PI) {
				rad -= 2 * Math.PI;
			} else if (rad < -Math.PI) {
				rad += 2 * Math.PI;
			}
			return rad * 180.0 / Math.PI;
		};
		
		// Run the init function to initialize values
		init();
	}
	
})(jQuery);