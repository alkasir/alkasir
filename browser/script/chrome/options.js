"use strict";

// imported for side effects only
import "./initUI";


// // Saves options to chrome.storage
// function save_options() {
//   var clientURL = document.getElementById('clientURL').value;

//   chrome.storage.local.set({
//       clientURL: clientURL
//   }, () => {
//       console.log("saved");
//   });
// }

// function restore_options() {
//   chrome.storage.local.get({
//       clientURL: 'http://localhost:8899',
//   }, (items) => {
//       document.getElementById('clientURL').value = items.clientURL;

//   });
// }
// document.addEventListener('DOMContentLoaded', restore_options);
// document.getElementById('save').addEventListener('click', save_options);
