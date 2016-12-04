// Progress modal works in combination with Twitter Bootstrap. It displays a
// modal containing a progress bar. The progress bar moves with a fixed speed,
// and is only removed from the view when the Finished() method is fired.
//
// It is suitable for events where the user has to wait for reply from the
// server.
;(function($) {

	$.progressmodal = function(el, options) {

		// Default settings values
		var defaults = {
			duration: 10*1000,
			updateInterval: 100,
		};

		var plugin = this;
		plugin.modal = el;
		plugin.progressbar = el.find(".progress .bar");

		// Plugin initialization
		var init = function() {
			plugin.settings = $.extend({}, defaults, options);

			// Set up the modal
			$(plugin.modal).modal({
				keyboard: false,
				backdrop: "static",
				show: false,
			});
		};

		var update = function() {
			plugin.progressbar.css("width", "" + getProgress() * 100 + "%");
		};

		// Run the progress bar
		plugin.Run = function(description, duration) {

			if (typeof description == "string") {
				plugin.modal.find("h3").html(description);
			}

			if (typeof duration == "number" && duration > 0) {
				plugin.settings.duration = duration;
			}

			// Show the modal
			plugin.modal.modal('show');

			// Set the start time
			plugin.startTime = (new Date()).getTime();

			plugin.interval = setInterval(update, plugin.settings.updateInterval);
		};



		// Determine how far (in decimal) we have come.
		var getProgress = function() {
			return Math.min(((new Date()).getTime() - plugin.startTime) / plugin.settings.duration, 1);
		};

		plugin.Finished = function(callback) {

			clearInterval(plugin.interval);
			plugin.progressbar.css("width", "100%");

			setTimeout(function() {
				
				plugin.modal.modal("hide");
				if (typeof callback == "function") {
					callback();
				}

			}, 1000);

		};

		// Run the init function
		init();
	}

})(jQuery);