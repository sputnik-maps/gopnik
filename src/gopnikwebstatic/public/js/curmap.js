$(function() {
  var ZoomControl = L.Control.extend({
    options: {
      position: 'bottomleft'
    },
    onAdd: function (map) {
      // create the control container with a particular class name
      var container = L.DomUtil.create('div', 'lf-zoom-control');
      container.innerHTML = map.getZoom();
      map.on('zoomend', function(ev) {
        container.innerHTML = map.getZoom();
      });
      return container;
    }
  });

  var map = new L.map('map', {
    center: [55.889, 37.594],
    zoom: 13
  });
  L.tileLayer('/tiles/{z}/{x}/{y}.png').addTo(map);
  map.addControl(new ZoomControl());
})
