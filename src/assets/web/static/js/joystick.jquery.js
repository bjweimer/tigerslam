/*!
 * joystick-js  v0.0.1 alpha
 * https://github.com/mattes/joystick-js
 *
 * Requires:
 * jQuery http://jquery.com
 * Raphaël http://raphaeljs.com
 *
 * Copyright 2012 Matthias Kadenbach
 * Released under the MIT license
 */
"use strict";
(function($){

  var methods = {
    // public methods

    init : function(options, positionCallback){

      var settings = $.extend({

        width:300, // px
        height:300, // px
        axis:'xy', // xy, x or y
        virtualPositionEasing:"easeInCirc", // value from _easingFormulas (see this file below)
        virtualPositionMaxValue:100,
        invert:false,
        resetPositionIfDirectionChanges:false,
        // isMobile: /Android|webOS|iPhone|iPad|iPod|BlackBerry|PlayBook|Kindle|Windows Phone/i.test(navigator.userAgent), // @info not used atm
        pathColor:"#ffffff"

      }, options);

      return this.each(function(){
        var $div = $(this), div = this;
        
        // @todo implement keyboard control

        $div.css({"width":settings.width, "height":settings.height});
        $div.css({"width":settings.width, "height":settings.height});
        $div.css("cursor", "pointer");

        var paper = Raphael(div, settings.width, settings.height);

        var path = paper.path().hide();
        path.attr({"stroke":settings.pathColor, "stroke-width":3, "arrow-end":"oval-wide-long"});

        var touch = {};
        var divOffset = $div.offset();

        var touchStart = function(e){
          e.preventDefault();

          if(typeof e.pageX == "undefined"){
            // @todo jquery doesnt generate pageX/Y for mobile clients. bug?
        	// for mobile devices (added by Mikael berg)
              if(typeof e.originalEvent.targetTouches != "undefined") {
              	e.pageX = e.originalEvent.targetTouches[0].pageX;
              	e.pageY = e.originalEvent.targetTouches[0].pageY;
              } else {
  	        	e.pageX = e.originalEvent.pageX;
  	            e.pageY = e.originalEvent.pageY;
              }
          }

          divOffset = $div.offset();
          
          $("body").attr("onmousedown", "return false;"); // @todo using <body onmousedown="..."> will lead to problems! (see below)

          $("body").bind("mousemove touchmove", touchMove);
          $("body").bind("mouseup touchend", touchEnd);

          $("body").css("cursor", "move");
          $div.css("cursor", "move");

          touch = {}; // new touch event
          touch.zero = {"x":e.pageX - divOffset.left, "y":e.pageY - divOffset.top};
          touch.new = {"x":0, "y":0};
          touch.delta = {"x":0, "y":0};
          touch.oldDelta = {"x":0, "y":0};
        }

        var touchMove = function(e){
          e.preventDefault();
          
          var activeField = {"top":0, "right":settings.width, "bottom":settings.height, "left":0, "centerX": settings.width/2, "centerY": settings.height/2};

          if(typeof e.pageX == "undefined"){
            // @todo jquery doesnt generate pageX/Y for mobile clients. bug?
        	// for mobile devices (added by Mikael berg)
            if(typeof e.originalEvent.targetTouches != "undefined") {
            	e.pageX = e.originalEvent.targetTouches[0].pageX;
            	e.pageY = e.originalEvent.targetTouches[0].pageY;
            } else {
	        	e.pageX = e.originalEvent.pageX;
	            e.pageY = e.originalEvent.pageY;
            }
          }

          // calculate position of new touch
          touch.new.x = _valueBetween(e.pageX - divOffset.left, activeField.left, activeField.right);
          touch.new.y = _valueBetween(e.pageY - divOffset.top, activeField.top, activeField.bottom);

          // save old delta
          touch.oldDelta.x = touch.delta.x;
          touch.oldDelta.y = touch.delta.y;
          
          // calculate new delta
          touch.delta.x = touch.new.x - touch.zero.x;
          touch.delta.y = touch.new.y - touch.zero.y;

          // reset to zero as soon as direction changes
          if(settings.resetPositionIfDirectionChanges){
            if((touch.zero.x < touch.new.x && touch.oldDelta.x > touch.delta.x) // right -> left
              || (touch.zero.x > touch.new.x && touch.oldDelta.x < touch.delta.x) // left -> right
              || (touch.zero.y < touch.new.y && touch.oldDelta.y > touch.delta.y) // down -> up
              || (touch.zero.y > touch.new.y && touch.oldDelta.y < touch.delta.y)){ // up -> down
              touch.zero = {"x":touch.new.x, "y":touch.new.y};
              touch.delta = {"x":0, "y":0};
              touch.oldDelta = {"x":0, "y":0};
            } 
          } 

          // calculate virtual position
          var vx = (_easing(_valueBetween(touch.delta.x / activeField.right, -1, 1), settings.virtualPositionEasing) *settings.virtualPositionMaxValue);
          var vy = (_easing(_valueBetween(touch.delta.y / activeField.bottom, -1, 1), settings.virtualPositionEasing) *settings.virtualPositionMaxValue);

          // invert?
          if(settings.invert){
            vx = -vx;
            vy = -vy;            
          }

          // animate path and restrict axis
          if(settings.axis == "x"){
            path.attr("path", "M" + touch.zero.x + "," + activeField.centerY + "L" + touch.new.x + "," + activeField.centerY + "z").show();
            vy = 0;
          } else if(settings.axis == "y"){
            path.attr("path", "M" + activeField.centerX + "," + touch.zero.y + "L" + activeField.centerX + "," + touch.new.y + "z").show();
            vx = 0;
          } else {
            path.attr("path", "M" + touch.zero.x + "," + touch.zero.y + "L" + touch.new.x + "," + touch.new.y + "z").show();
          } 

          // callback ...
          if(positionCallback) positionCallback.call(this, vx, vy, false);

        }

        var touchEnd = function(e){
          e.preventDefault();
          
          $("body").unbind("mousemove touchmove");
          $("body").unbind("mouseup touchend");

          // callback ...
          if(positionCallback) positionCallback.call(this, 0, 0, false);

          $("body").css("cursor", "");           
          $("body").removeAttr("onmousedown"); // @todo using <body onmousedown="..."> will lead to problems!

          path.hide();    

          $div.css("cursor", "pointer"); 
        }


        $div.bind("mousedown touchstart", touchStart);

      });

    }
  };


  // private methods

  // return value between min and max
  function _valueBetween(value, min, max){
    if(value < min){
      return min;
    } else if(value > max){
      return max;
    }
    else {
      return value;
    }
  }

  // easing n (-1..0 or 0..1) with type
  function _easing(n, type){
    if(!_easingFormulas[type]){
      $.error("Easing formula " + type + " does not exist on jQuery.joystick");
      return false;
    }

    if(n >= 0){
      return _easingFormulas[type].call(this, n);
    } else {
      return _easingFormulas[type].call(this, Math.abs(n))*-1;
    }
  }

  var _easingFormulas = {
    // more https://github.com/jeremyckahn/shifty/blob/master/src/shifty.formulas.js
    // and http://rekapi.com/ease.html
    linear: function(pos) {
        return pos;
    },
    easeInCirc: function(pos){
      return -(Math.sqrt(1 - (pos*pos)) - 1);
    },
    easeOutCirc: function(pos){
      return Math.sqrt(1 - Math.pow((pos-1), 2))
    }     
  }


  // method dispatcher
  $.fn.joystick = function(method){
    if (methods[method]) {
      return methods[method].apply( this, Array.prototype.slice.call(arguments, 1));
    } else if (typeof method === "object" || !method){
      return methods.init.apply(this, arguments);
    } else {
      $.error("Method " +  method + " does not exist on jQuery.joystick");
    }    
  };

})(jQuery);