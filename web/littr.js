// LIT library
;(function () {
  'use strict'

  // LIT object
  window.LIT = {}
  window.LIT.event = null
  window.LIT.version = 'LittrJS v0.8.0'

  // onload event listener
  addEventListener('load', event => {
    console.log(LIT.version)
  })
})()

// Add Umami analytics - https://umami.is
const host = window.location.hostname;
if (host != "localhost") {
   window.onload = () => {
      var x = document.createElement('script')
      x.setAttribute('src', 'https://umami.vxn.dev/script.js')
      x.setAttribute('data-website-id', '23275649-ae96-43b8-9362-93af740f6560')
      document.body.appendChild(x)
   }
};

