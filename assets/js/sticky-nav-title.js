/* eslint-env browser */

/**
 * Nav/Title replacement
 * Used on invividual post pages, displays the post title in place of the nav
 * bar when scrolling past the title
 *
 * Usage:
 * ```
 * Casper.stickyTitle({
 *     navSelector: '.site-nav-main',
 *     titleSelector: '.post-full-title',
 *     activeClass: 'nav-post-title-active'
 * });
 * ```
 */

(function (window, document) {
    // set up Casper as a global object
    if (!window.Casper) {
        window.Casper = {};
    }

    window.Casper.stickyNavTitle = function stickyNavTitle(options) {
        var nav = document.querySelector(options.navSelector);
        var title = document.querySelector(options.titleSelector);

        var lastScrollY = window.scrollY;
        var ticking = false;

        function onScroll() {
            lastScrollY = window.scrollY;
            requestTick();
        }

        function requestTick() {
            if (!ticking) {
                requestAnimationFrame(update);
            }
            ticking = true;
        }

        function update() {
            var trigger = title.getBoundingClientRect().top + window.scrollY;
            var triggerOffset = title.offsetHeight + 35;

            // show/hide post title
            if (lastScrollY >= trigger + triggerOffset) {
                nav.classList.add(options.activeClass);
            } else {
                nav.classList.remove(options.activeClass);
            }

            ticking = false;
        }

        window.addEventListener('scroll', onScroll, {passive: true});

        update();
    };
    
})(window, document);
