const site = window.location.hostname;
console.log("Site:", site);

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

    const addressWrapper = document.querySelector('.styles__AddressWrapper-fshdp-8-111-1__sc-13x5vko-0.jDtXfP');
    const addressElement = addressWrapper.childNodes[0];
    let listingTitle = addressElement ? addressElement.textContent.trim() : null;
    if (message.action == "get_listing_title") {
        sendResponse({title: listingTitle});
    }

    // Since `fetch` is asynchronous, must send an explicit `true`
    return true;
}

chrome.runtime.onMessage.addListener(handleMessages);