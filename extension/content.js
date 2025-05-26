const site = window.location.hostname;
console.log("Site:", site);

console.log("Content script loaded");
alert("Content script loaded");
InsertHtml();

// Wait for the DOM to be fully loaded
document.addEventListener('DOMContentLoaded', function () {
    console.log("DOM fully loaded and parsed");
    alert("Dom loaded");
    InsertHtml();
});

function InsertHtml() {
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
}