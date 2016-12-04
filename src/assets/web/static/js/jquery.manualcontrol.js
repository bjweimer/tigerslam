;(function($) {

	$.manualcontrol = function(el, options) {

		// Default settings values
		var defaults = {
			joystick: {
				width: 300,
				height: 300,
				pathColor: "#f00",
				virtualPositionMaxValue: 10.0,
			},
			isConnected: false,
			sendInterval: 500,
			turnAlpha: 0.3,
		};

		var plugin = this;

		var init = function() {
			plugin.settings = $.extend({}, defaults, options);
			plugin.el = el;

			// Joystick saves information to these variables
			plugin.joyX = 0;
			plugin.joyY = 0;

			// Will later be used as interval variable
			plugin.sendInterval = false;

			enableJoystick();
			startSpeedSetting();
		};

		plugin.SetSpeeds = function(left, right) {
			$.ajax("/api/set/motor/speeds", {
				dataType: 'json',
				data: {
					left: left,
					right: right,
				},
				error: speedError,
			});
		};

		// What should happen if the server returns an error from a speed
		// update.
		var speedError = function(jqXHR) {
			stopSpeedSetting();

			if (jqXHR.responseText) {
				$("#trackpad").html("Error: " + jqXHR.responseText);
			} else {
				$("#trackpad").html("Error: Lost connection?");
			}
		}

		// Set up the joystick plugin
		var enableJoystick = function() {
			plugin.joystick = plugin.el.find("#joystick").joystick( plugin.settings.joystick, 
				function(x, y) {
					plugin.joyX = x;
					plugin.joyY = -y;
				});
		}

		// Disable the joystick plugin
		var disableJoystick = function() {
			plugin.joystick.remove();
		};

		var startSpeedSetting = function() {
			plugin.sendInterval = setInterval(setSpeedsFromJoystick, plugin.settings.sendInterval);
		};

		var stopSpeedSetting = function() {
			clearInterval(plugin.sendInterval);
		};

		var setSpeedsFromJoystick = function() {
			var speed = Math.min(Math.max(plugin.joyY, -1.0), 1.0);
			var turn = Math.min(Math.max(plugin.joyX, -1.0), 1.0);

			var left = speed + plugin.settings.turnAlpha * turn;
			var right = speed - plugin.settings.turnAlpha * turn;

			var normalizer = Math.max(left, right);
			if (normalizer > 1.0) {
				left = left / normalizer;
				right = right / normalizer;
			}

			plugin.SetSpeeds(left, right);
		};

		// Run the init function
		init();
	};

})(jQuery);