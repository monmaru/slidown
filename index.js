'use strict';

$(function() {
  $('#download').on('click', onDownloadClicked);
});

function onDownloadClicked() {
  var url = $('#url').val();
  if (!url || url === '') {
    Materialize.toast('スライドのURLを入力してください！！', 3000);
  } else {
    download(url);
  }
}

function download(url) {
  var $loading = $('#loading');
  $loading.show();
  var xhr = new XMLHttpRequest();
  xhr.open("POST", '/api/slideshare/download', true);
  xhr.responseType = "blob";
  xhr.onload = function (oEvent) {
    try {
      if (this.status == 200) {
        var blob = xhr.response;
        var fileName = xhr.getResponseHeader("X-FileName")
        if (window.navigator.msSaveBlob) {
          window.navigator.msSaveBlob(blob, fileName);
        }
        else {
          var objectURL = window.URL.createObjectURL(blob);
          var link = document.createElement("a");
          document.body.appendChild(link);
          link.href = objectURL;
          link.download = fileName;
          link.click();
          document.body.removeChild(link);
        }
      } else {
        alert('ERRORS!!!');
      }
    } finally {
      $loading.hide();
    }
  };
  xhr.setRequestHeader("Content-Type", "application/json");
  xhr.send(JSON.stringify({'url': url}));
}
