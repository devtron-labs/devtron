/* eslint-env browser */

/**
 * Gallery card support
 * Used on any individual post/page
 *
 * Detects when a gallery card has been used and applies sizing to make sure
 * the display matches what is seen in the editor.
 */

(function (window, document) {
    var resizeImagesInGalleries = function resizeImagesInGalleries() {
        var images = document.querySelectorAll('.kg-gallery-image img');
        images.forEach(function (image) {
            var container = image.closest('.kg-gallery-image');
            var width = image.attributes.width.value;
            var height = image.attributes.height.value;
            var ratio = width / height;
            container.style.flex = ratio + ' 1 0%';
        });
    };

    document.addEventListener('DOMContentLoaded', resizeImagesInGalleries);
   
  

})(window, document);
$(document).ready(function() { 
	$("a[href^='http']").attr("target","_blank");
  });
