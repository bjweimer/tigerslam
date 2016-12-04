/*
 * Options for SLAM Map Type, which specifies the appropriate options for an
 * image map type.
 * 
 * Specifies the URL to load maps from, the tile size and zoom levels.
 */

function getTileUrl(coord, zoom) {
	var normalizedCoord = getNormalizedCoord(coord, zoom);
    if (!normalizedCoord) {
            return null
    }
    return "/api/get/slam/image/tile/?zoomLevel=" + zoom + "&tileX=" + 
                    coord.x + "&tileY=" + coord.y; // + "&timestamp=" + (new Date()).getTime();
}

var slamTypeOptions = {
        getTileUrl: getTileUrl,
        tileSize: new google.maps.Size(256, 256),
        maxZoom: 7,
        minZoom: 1,
        name: "SLAM",
}

function EuclideanProjection(worldSize) {
	this.worldSize = worldSize;
    var EUCLIDEAN_RANGE = 256;
    this.pixelOrigin_ = new google.maps.Point(EUCLIDEAN_RANGE / 2, EUCLIDEAN_RANGE / 2);
    this.pixelsPerLonDegree_ = EUCLIDEAN_RANGE / 360;
    this.pixelsPerLonRadian_ = EUCLIDEAN_RANGE / (2 * Math.PI);
    this.scaleLat = 2;		// Height - multiplication scale factor
    this.scaleLng = 1;		// Width - multiplication scale factor
    this.offsetLat = 0;		// Height - direct offset +/-
    this.offsetLng = 0;		// Width - direct offset +/-
};

EuclideanProjection.prototype.fromLatLngToPoint = function(latLng, opt_point) {
    var point = opt_point || new google.maps.Point(0, 0);

    var origin = this.pixelOrigin_;
    point.x = (origin.x + (latLng.lng() + this.offsetLng ) * this.scaleLng * this.pixelsPerLonDegree_);
    // NOTE(appleton): Truncating to 0.9999 effectively limits latitude to
    // 89.189.  This is about a third of a tile past the edge of the world tile.
    point.y = (origin.y + (latLng.lat() + this.offsetLat ) * this.scaleLat * this.pixelsPerLonDegree_);
    return point;
};

EuclideanProjection.prototype.fromPointToLatLng = function(point) {
    var me = this;

    var origin = me.pixelOrigin_;
    var lng = (((point.x - origin.x) / me.pixelsPerLonDegree_) / this.scaleLng) - this.offsetLng;
    var lat = (((point.y - origin.y) / me.pixelsPerLonDegree_) / this.scaleLat) - this.offsetLat;
    return new google.maps.LatLng(lat , lng, true);
};

/*
 * Translates from physical coordinates
 */
EuclideanProjection.prototype.fromPhysicalToPoint = function(x, y) {
	return new google.maps.Point(x * 256 / this.worldSize, (this.worldSize - y) * 256 / this.worldSize);
}

/*
 * Translates to physical coordinates
 */
EuclideanProjection.prototype.fromPointToPhysical = function(point) {
	return {
		x: point.x * this.worldSize / 256,
		y: this.worldSize - point.y * this.worldSize / 256,
	}
}

/*
 * Translates from physical coordinates to latlng
 */
EuclideanProjection.prototype.fromPhysicalToLatLng = function(x, y, noWrap) {
	var point = this.fromPhysicalToPoint(x, y)
	return this.fromPointToLatLng(point, noWrap)
}

var slamMapType = new google.maps.ImageMapType(slamTypeOptions);

slamMapType.tiles = {};

slamMapType._getTile = slamMapType.getTile;
slamMapType.getTile = function(coord, zoom, ownerDocument) {
	var tile = slamMapType._getTile(coord, zoom, ownerDocument);
	
	var tileID = "x" + coord.x + "y" + coord.y + "z" + zoom;
	// var tileID = Math.random().toString();
	tile.tileID = tileID;
	tile.coord = coord;
	tile.zoom = zoom;
	tile.ownerDocument = ownerDocument;
	slamMapType.tiles[tileID] = tile;
	
	return tile;
}

slamMapType._releaseTile = slamMapType.releaseTile;
slamMapType.releaseTile = function(tile) {
	this._releaseTile(tile)
	delete this.tiles[tile.tileID];
}

slamMapType.updateTile = function(tile) {
	var tileUrl = getTileUrl(tile.coord, tile.zoom);
	if (!tileUrl) {
		return
	}
	
	$(tile).find("img").attr("src", tileUrl);

	// tile.innerHTML = '<img style="width: 256px; height: 256px; -webkit-user-select: none; border: 0px; padding: 0px; margin: 0px; -webkit-transform: translateZ(0);" src="' + tileUrl + '" draggable="false">';
	// var img = new Image();
	// img.onload = function() {
	// 	tile.innerHTML = '<img src="' + tileUrl + '">';
	// 	tile.replaceChild(img, tile.firstChild);
	// 	img.onload = null;
	// 	delete img;
	// };
	// img.src = tileUrl;
};

slamMapType.update = function() {
	for (var tileID in this.tiles) {
		this.updateTile(this.tiles[tileID]);
	}
}

function createSlamMap(mapSizeMeters) {
	var mapOptions = {
		zoom: 2,
		center: new google.maps.LatLng(0, 0),
		streetViewControl: false,
		mapTypeControlOptions: {
			mapTypeIds: []
		}
	};
	
	slamMapType.projection = new EuclideanProjection(mapSizeMeters);
	
	var map = new google.maps.Map(document.getElementById('map_canvas'),
			mapOptions);
	
	map.mapTypes.set('slam', slamMapType);
	map.setMapTypeId('slam');
	
	return map
}

//Normalizes the coords that tiles repeat across the x axis (horizontally)
//like the standard Google map tiles.
function getNormalizedCoord(coord, zoom) {
	var y = coord.y;
	var x = coord.x;

	// tile range in one direction range is dependent on zoom level
	// 0 = 1 tile, 1 = 2 tiles, 2 = 4 tiles, 3 = 8 tiles, etc
	var tileRange = 1 << zoom;
	
	// dont't repeat in any axis
	if (x < 0 || x >= tileRange || y < 0 || y >= tileRange) {
		return null;
	}

	return {
		x: x,
		y: y
	};
}

