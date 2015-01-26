$.getJSON( "/heattiles_zooms", function( data ) {
  var tiles = new L.TileLayer('http://{s}.tile.osm.org/{z}/{x}/{y}.png', {
    attribution: 'Map data &copy; <a href="http://openstreetmap.org">OpenStreetMap</a> contributors, <a href="http://creativecommons.org/licenses/by-sa/2.0/">CC-BY-SA</a>',
    minZoom: 3,
    maxZoom: 18
  });
  var map = new L.map('map', {
    center: [55.889, 37.594],
    zoom: 13,
    layers: [tiles]
  });
  var baseLayers = {
    "OSM": tiles
  };

  var overlays = {};
  for(var i = data.Min; i <= data.Max; i++) {
    var heatTiles = new L.TileLayer('/heattiles/' + i + '/{z}/{x}/{y}.png', {
      minZoom: 3,
      maxZoom: 18
    });
    var heatTilesKey = 'Times (zoom = ' + i + ')';
    overlays[heatTilesKey] = heatTiles
  }

  L.control.layers(baseLayers, overlays).addTo(map);
});
