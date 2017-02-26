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

  $.ajax({
    type: 'POST',
    url: '/api/slideshare/download', // TODO
    dataType: 'json',
    data: JSON.stringify({'url': url}),
  }).always(function() {
    $loading.hide();
  }).done(function(data, textStatus, jqXHR) {
    var fileName = jqXHR.getResponseHeader('X-FileName');
    console.log(fileName);
    var blob = new Blob([data], {type: "application/octet-stream"});
    
    if (window.navigator.msSaveBlob) {
      window.navigator.msSaveBlob(blob, fileName); // IE用
    } else {
      var downloadUrl  = (window.URL || window.webkitURL).createObjectURL(blob);
      var link = document.createElement('a');
      link.href = downloadUrl;
      link.download = fileName;
      link.click();
      (window.URL || window.webkitURL).revokeObjectURL(downloadUrl);
    }
  }).fail(function(jqXHR, textStatus, errorThrown) {
    alert('ERRORS: ' + textStatus + ' ' + errorThrown);
  });
}
