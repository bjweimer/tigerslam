;(function($) {

    $.log = function(el, options) {

        var defaults = {
            fullUrl: "/api/get/log/",
            streamUrl: "/api/streaming/log/"
        }

        var plugin = this;

        plugin.settings = {}

        var init = function() {
            plugin.settings = $.extend({}, defaults, options);
            plugin.el = el;

            plugin.socket = false;
        }


        plugin.Start = function() {
        	fullUpdate();

        	// Set up web socket
//        	plugin.socket = $.Websocket("ws://localhost:8000/api/streaming/log/", {
//        		events: {
//        			logEntry: function(e) {
//        				plugin.el.prepend(e.data);
//        			}
//        		}
//        	});
        }

        plugin.Stop = function() {
        }

        var fullUpdate = function() {
            $.getJSON(plugin.settings.fullUrl, function(data) {
            	var str = '';
            	for (var i = data.length - 1; i >= 0; i--) {
            		str += data[i];
            	}
            	plugin.el.html(str);
            })
        }

        init();

    }

})(jQuery);
