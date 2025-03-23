// LIT library
;(function () {
  'use strict'

  // LIT object
  window.LIT = {}
  window.LIT.event = null
  window.LIT.version = 'LittrJS v0.7.0'

  // feature detection: mobile device
  if ('ontouchstart' in window || (window.DocumentTouch && document instanceof DocumentTouch)) {
    // feature detection: share
    if (navigator.share) {
      if (typeof window.WebShare !== 'function')
        window.WebShare = function (text, url, title) {
          url =
            url ||
            (document.querySelector('link[rel=canonical]')
              ? document.querySelector('link[rel=canonical]').href
              : document.location.href)
          title = title || document.title
          text = text || document.title
          navigator
            .share({
              title: title,
              text: text,
              url: url
            })
            .catch(console.error)
        }
    }
  }

  // fix links and images
  window.LIT.fixLinks = function () {
    if ($('#table-flow article span:not(.ff)').length) {
      $('#table-flow article span:not(.ff)').each(function () {
        $(this)
          .html(
            $(this)
              .html()
              .replace(
                /(https:\/\/[\w?=&.\/-;#~%-]+(?![\w\s?&.\/;#~%"=-]*>))/g,
                '<a class="red-text" target=_blank href="$1">$1</a> '
              )
          )
          .addClass('ff')
      
        $(this)
          .html(
            $(this)
              .html()
              .replace(
                /#([\w]+)/g,
                '<a class="red-text" target=_blank href="/flow/hashtags/$1">#$1</a> '
              )
          )
          .addClass('ff')

        $(this)
          .html(
            $(this)
              .html()
              .replace(
                /@([\w]+)/g,
                '<a class="red-text" target=_blank href="/flow/users/$1">@$1</a> '
              )
          )
          .addClass('ff')
      })

      $('#table-flow a').each(function () {
        let x = $(this).html()
        if (x.endsWith('.webp') || x.endsWith('.jpg') || x.endsWith('.jpeg') || x.endsWith('.png')) {
          let u = $(this).attr('href')
          $(this)
            .html('<img class="ff" width=25% src="' + u + '">')
            .addClass('ff')
        }
      })
    }
    $('a>img').parent().attr('href', '')
  }

  // fix UI glitches
  window.LIT.fixUI = function () {
    // fix links and images
    LIT.fixLinks()

    // fix cursors
    $('#table-users p.bold').css('cursor', 'pointer')
    // set some tables sortable
    $('#table-stats-flow,#table-users,#table-poll').addClass('sortable')

    // test 4 UI fix done
    if ($('main').data('fixedUI')) return false
    $('main').data('fixedUI', true)

    // fix tables bottom padding
    $('table').css('padding-bottom', '2rem')

    // STATS tab
    if ($('#table-stats-flow') && $('#table-stats-flow').length) {
      $('#nav-bottom > a:nth-child(1)').click(function () {
        LIT.scrollTop()
      })
    }
    // USERS tab
    if ($('#table-users') && $('#table-users').length) {
      $('#nav-bottom > a:nth-child(2)').click(function () {
        LIT.scrollTop()
      })
    }
    // POLLS tab
    if ($('#table-poll') && $('#table-poll').length) {
      $('#nav-bottom > a:nth-child(4)').click(function () {
        LIT.scrollTop()
      })
    }
    // FLOW tab
    if ($('#table-flow') && $('#table-flow').length) {
      $('#nav-bottom > a:nth-child(5)').click(function () {
        LIT.scrollTop()
        //location("/flow")
      })
    }

    // WebShare click event
    if (typeof window.WebShare === 'function') {
      $('#nav-top > dialog > img').click(function () {
        window.WebShare()
      })
    }
  }

  // onload event listener
  addEventListener('load', event => {
    console.log(LIT.version)

    // set fix UI action interval
    setInterval(LIT.fixUI, 250)
  })
})()

// add Umami analytics - https://umami.is
const host = window.location.hostname;
if (host != "localhost") {
   window.onload = () => {
      var x = document.createElement('script')
      x.setAttribute('src', 'https://umami.vxn.dev/script.js')
      x.setAttribute('data-website-id', '23275649-ae96-43b8-9362-93af740f6560')
      document.body.appendChild(x)
   }
};

