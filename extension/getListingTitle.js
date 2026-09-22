const domainName = window.location.hostname;
console.log("Site:", domainName);

console.log("Get listing title script loaded");
//alert("Content script loaded");

/* function InsertHtml() {
    alert("InsertHtml function called");
    // Find the first element with the class 'layout-static-column-container'
    const container = document.querySelector('.layout-static-column-container');
    console.log("Found container:\n\n\n\n\n\n",container);

    if (container && container.parentNode) {
        // Create a new header element
        const header = document.createElement('h1');
        header.textContent = 'hello world';
        // Insert the header before the container in the DOM
        container.parentNode.insertBefore(header, container);
    }
} */

// Event listener
function handleMessages(message, sender, sendResponse) {
  //console.log("Getting message...");

  console.log("Recieved message...");

  const houseRegex = RegExp("^https:\/\/www\.zillow\.com\/homedetails\/.*");
  const apartmentRegex = RegExp("^https:\/\/www\.zillow\.com\/apartments\/.*$");
  const listingUrl = location.href;

  let listingTitle;
  let listingType;

  if (houseRegex.test(listingUrl)) {
    listingTitle = document.title.split("|")[0].trim();
    listingType = "house";
  } else if (apartmentRegex.test(listingUrl)) {
    listingTitle = document.title.split("|")[0].trim();
    listingType = "apartment";
  } else {
    console.log("Listing was neither for a house nor apartment.");
  }

  console.log(
    "Listing title, listing type: " + listingTitle + ", " + listingType,
  );

  if (message.action != "get_listing_title") {
    sendResponse({ error: "function not supported" });
    return false;
  }

  if (listingTitle && listingType) {
    sendResponse({ title: listingTitle, type: listingType });
  } else {
    sendResponse({ error: "listing not house nor apartment" });
  }

  // Since `fetch` is asynchronous, must send an explicit `true`
  return true;
}

chrome.runtime.onMessage.addListener(handleMessages);
