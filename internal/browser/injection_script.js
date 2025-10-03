// Safari WebDriver Injection Script
// This script is automatically injected into every page

(function() {
  'use strict';
  
  // Mark that injection has occurred
  window.__webdriverInjected = true;
  
  // Add helper utilities that can be used by the automation
  window.__webdriverHelpers = {
    // Get element information
    getElementInfo: function(element) {
      if (!element) return null;
      return {
        tagName: element.tagName,
        id: element.id,
        className: element.className,
        textContent: element.textContent,
        value: element.value
      };
    },
    
    // Wait for selector to appear
    waitForSelector: function(selector, timeout) {
      timeout = timeout || 30000;
      var start = Date.now();
      return new Promise(function(resolve, reject) {
        var check = function() {
          var element = document.querySelector(selector);
          if (element) {
            resolve(element);
          } else if (Date.now() - start >= timeout) {
            reject(new Error('Timeout waiting for selector: ' + selector));
          } else {
            setTimeout(check, 100);
          }
        };
        check();
      });
    },
    
    // Check if element is visible
    isVisible: function(element) {
      if (!element) return false;
      var style = window.getComputedStyle(element);
      return style.display !== 'none' && 
             style.visibility !== 'hidden' && 
             style.opacity !== '0' &&
             element.offsetWidth > 0 &&
             element.offsetHeight > 0;
    },
    
    // Get page metrics
    getPageMetrics: function() {
      return {
        url: window.location.href,
        title: document.title,
        readyState: document.readyState,
        timestamp: Date.now(),
        viewport: {
          width: window.innerWidth,
          height: window.innerHeight
        },
        scroll: {
          x: window.scrollX,
          y: window.scrollY
        }
      };
    }
  };
  
  console.log('[WebDriver] Injection script loaded');
})();

