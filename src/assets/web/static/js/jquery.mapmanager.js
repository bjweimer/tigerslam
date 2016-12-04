// Map Manager is responsible for opening maps
;(function($) {

	$.mapmanager = function(options) {

		// Default settings values
		var defaults = {
			mapOpenTime: 100*1000,
			newMapTime: 5*1000,
			storedMapTime: 100*1000,
			progressModal: $("#progress-modal"),
		};

		var plugin = this;

		// Plugin initialization
		var init = function() {
			plugin.settings = $.extend({}, defaults, options);
			plugin.detailsModal = $("#map-details-modal").modal({show: false});

			// Set up the progress modal
			plugin.progressModal = new $.progressmodal(plugin.settings.progressModal, {
				duration: plugin.settings.mapOpenTime,
			});

			// Set up new map links
			$("[data-role=start-new-map]").click(function() {
				plugin.StartNewMap($(this).attr("data-algorithm"));
			})

			// Set up details-modals
			$("[data-role=open-details]").click(function() {
				plugin.OpenDetails($(this).attr("data-filename"));
			})
		};

		// Open details
		// Open the modal and get hold of the data (from server), fill inn the
		// data.
		plugin.OpenDetails = function(filename) {

			var details = {};

			// Send AJAX request for the map's meta data
			$.ajax("/api/get/mapstorage/metadata", {
				data: {
					filename: filename
				},
				dataType: 'json',
				beforeSend: function() { plugin.progressModal.Run("Fetching details ...", 1000); },
				success: function(data) {
					plugin.FillInDetailsModal(filename, data);
					plugin.progressModal.Finished(function() {
						plugin.detailsModal.modal("show");
					});
				},
				error: function() {
					plugin.progressModal.Finished();
					alert("Error while opening");
				}
			})
		};

		plugin.FillInDetailsModal = function(filename, data) {
			plugin.detailsModal.find("[data-role=map-name]").html(data.Name);
			plugin.detailsModal.find("#map-name").val(data.Name);
			plugin.detailsModal.find("[data-role=thumbnail]").attr("src", "/api/get/mapstorage/thumbnail?filename=" + filename);
			plugin.detailsModal.find("[data-role=description]").html(data.Description);

			// Size (resolution)
			plugin.detailsModal.find("[data-role=size-x]").html(data.Mdp.MapDimensions[0]);
			plugin.detailsModal.find("[data-role=size-y]").html(data.Mdp.MapDimensions[1]);

			// Lengths (physical)
			plugin.detailsModal.find("[data-role=length-x]").html(data.Mdp.MapDimensions[0] * data.Mdp.CellLength + " m");
			plugin.detailsModal.find("[data-role=length-y]").html(data.Mdp.MapDimensions[1] * data.Mdp.CellLength + " m");

			// Cell area
			plugin.detailsModal.find("[data-role=cell-area]").html("" + data.Mdp.CellLength + "<sup>2</sup> m");

			// Open-link
			plugin.detailsModal.find("[data-role=open-stored-map]").click(function() {
				plugin.StartFromStored(filename);
			});

			// Save map meta data button
			plugin.detailsModal.find("[data-role=save-map-metadata]").click(function() {
				plugin.SaveMetaData(filename);
			})
		};

		// Save (edited) meta data from the modal
		plugin.SaveMetaData = function(filename) {

			var name = plugin.detailsModal.find("#map-name").val();
			if (name == "") {
				alert("The map must have a name");
				return
			}

			var description = plugin.detailsModal.find("#map-description").val();
			console.log(description);

			plugin.progressModal.Run("Saving details ...", 1000);

			var saveDescription = function(data) {
				console.log(data);
				finishAndReload();
			};

			// Save name
			$.ajax("/api/set/mapstorage/mapname", {
				dataType: 'json',
				data: {
					filename: filename,
					newname: name
				},
				success: saveDescription,
				error: errorFunc, 
			});

			};

		// Start a completely new map.
		plugin.StartNewMap = function(algorithm) {

			// Start the progress bar
			plugin.progressModal.Run("Please wait while " + algorithm + " is initialized ...", plugin.settings.newMapTime);

			// Send request to server
			$.ajax("/api/set/slam/initialize", {
				data: {
					'algorithm': algorithm
				},
				error: errorFunc,
				success: finishAndReload,
			});

		};

		// Start map from saved map
		plugin.StartFromStored = function(filename) {
			
			// Close details modal
			plugin.detailsModal.modal("hide");

			// Start the progress bar
			plugin.progressModal.Run("Please wait while opening " + filename, plugin.settings.storedMapTime);

			$.ajax("/api/set/slam/initialize-from-stored-map", {
				data: {
					algorithm: 'hectorslam',
					filename: filename,
				},
				error: errorFunc,
				success: finishAndReload,
			});

		};

		// What to do when a map is successfully loaded.
		var finishAndReload = function() {
			console.log("Finish and reload ..")
			plugin.progressModal.Finished();
			location.reload();
		}

		var errorFunc = function(jqXHR, textStatus, errorThrown) {
			plugin.progressmodal.Finished();
			alert("Error: " + jqXHR.responseText)
		}

		// Run the init function
		init();
	}

})(jQuery);