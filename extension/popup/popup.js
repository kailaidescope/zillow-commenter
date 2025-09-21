/* const API_ADDRESS = window.localStorage.getItem('zillow_commenter_api_address') || 'localhost';
const API_PORT = "3000";
const API_URL = `https://${API_ADDRESS}:${API_PORT}/api/v1`; */

let API_ADDRESS = window.localStorage.getItem('zillow_commenter_api_address') || "";
let API_URL = `${API_ADDRESS}/api/v1`;

// Get current tabs
const CURRENT_TABS = await chrome.tabs.query({active: true, lastFocusedWindow: true});

// Log the current user ID from localStorage
console.log("User ID: ", getLocalUserId());

// Set the user ID in localStorage if it doesn't exist
setUserId();

// Populate comments when the popup is opened
populateComments();

// Hide error field
hideErrorField()

// Handles the comment form submission
document.getElementById('comment-form').addEventListener('submit', handleCommentSubmission);

// Add hotkey for getting listing title
document.addEventListener("keyup", (event) => {
    if (event.altKey && event.key == "m") {
        getListingTitle();
    }
});

// Initialize html element listeners
initializeCommentInputValidation();
initializeZillowetteIcon();

console.log("Saved current tab ID:", CURRENT_TABS[0] ? CURRENT_TABS[0].id : "failed to save");

// Sets a unique user ID in localStorage if it doesn't exist
//
// Note: The localstorage persists between browser sessions, and incognito mode
function setUserId() {
    let userId = window.localStorage.getItem('zillow_commenter_user_id');
    if (!userId) {
        getNewUserId((result, error=null) => {
            if (error) {
                // If there's an error retrieving the user ID, log it
                console.error('Error retrieving user ID:', error);
                //setUserId();
                return;
            }
            if (result) {
                // If a user ID is found, parse it and log it
                const parsedResult = JSON.parse(result);
                console.log('Retrieved user ID:', parsedResult.user_id);
                window.localStorage.setItem('zillow_commenter_user_id', parsedResult.user_id);
            } else {
                // If no user ID is found, generate a new one
                console.log('No user ID found, generating a new one.');
                //setUserId();
            }
        });
        /* userId = "user_" + Math.random().toString(36).substring(2, 15);
        window.localStorage.setItem('zillow_commenter_user_id', userId); */
    }
}

// Retrieves the user ID from localStorage
function getLocalUserId() {
    return window.localStorage.getItem('zillow_commenter_user_id');
}


// Tab switching logic
document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', function() {
        // Remove active from all tabs and contents
        document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
        // Add active to clicked tab and corresponding content
        this.classList.add('active');
        document.getElementById(this.dataset.tab).classList.add('active');
    });
});

// Function to populate comments list in the DOM
async function populateComments() {
    // Get comments element from the DOM
    const commentsListElement = document.querySelector('.comments-list');
    //console.log('Populating comments.');
    if (!commentsListElement) return;

    // Clear existing comments
    commentsListElement.innerHTML = '';

    const listingIdAndType = await getListingIDAndType();

    if (!listingIdAndType) {
        displayComments(null);
        console.error("No valid listing ID found in the current URL.");
        return;
    }

    // Fetch comments from the API using the listing ID
    getCommentsByListingId(listingIdAndType, displayComments);
}

// Displays comments not found error
function displayCommentFetchError(error) {
    console.error('Error fetching comments:', error);
    const commentsListElement = document.querySelector('.comments-list');
    const li = document.createElement('li');
    li.innerHTML = '<strong>Error fetching comments.</strong> Please try again later.';
    
    // REFRESH BUTTON
    const refreshBtn = document.createElement('button');
    refreshBtn.textContent = 'Refresh';
    refreshBtn.style.display = 'block';
    refreshBtn.style.marginTop = '8px';

    // Refresh button re-populates comments upon click.
    refreshBtn.onclick = function() {
        populateComments();
    };

    // Format and compile error message
    li.appendChild(document.createElement('br'));
    li.appendChild(refreshBtn);
    commentsListElement.appendChild(li);
    const submitButton = document.querySelector('#comment-form button[type="submit"]');
    if (submitButton) {
        turnOffSubmitButton(submitButton);
    }
    return
}

// Function to display comments in the list
function displayComments(result, error=null) {
    // Autofill username input with saved username
    const usernameInput = document.getElementById('username-input');
    if (usernameInput) {
        const savedUsername = getSavedUsername();
        if (savedUsername) {
            usernameInput.value = savedUsername;
        }
    }

    // Get comments element from the DOM
    const commentsListElement = document.querySelector('.comments-list');
    if (!commentsListElement) return;
    
    // Clear existing comments
    commentsListElement.innerHTML = '';

    if (error) {
        displayCommentFetchError(error);
        return;
    }

    let comments = null;

    //console.log('Fetched comments:', result);
    if (result) {
        try {
            comments = JSON.parse(result);
            console.log(`Returned ${comments ? comments.length : 0} comments`);
            //console.log('Parsed comments:', comments);
        } catch (error) {
            console.error('Error parsing comments:', error);
            const li = document.createElement('li');
            li.textContent = 'Error loading comments.';
            commentsListElement.appendChild(li);
            return;
        }
    } else {
        // If no comments are returned, check if we have a valid listing ID
        getListingIDAndType().then(listingIdAndType => {
            if (!listingIdAndType) {
                // If no valid listing ID, disable the submit button and show an error message
                console.error("No valid listing ID or type found in the current URL.");
                const li = document.createElement('li');
                li.textContent = 'Not on an eligible zillow listing page.';
                commentsListElement.appendChild(li);
                return;
            } else {
                // If we have a valid listing ID but no comments, display a message
                const li = document.createElement('li');
                li.textContent = 'No comments found for this listing.';
                commentsListElement.appendChild(li);
            }
        });
        
        return;
    }

    // Check if there are any comments
    if (comments !== null) {
        console.log(`Displaying ${comments.length} comments` );
    } else {
        console.log('Displaying comment, comments is null');
    }

    /* if () {
        console.error('Invalid comments data:', comments);
        // If comments are bugged, display an error message
        const li = document.createElement('li');
        li.textContent = 'Our apologies, this type of listing does not yet support comments.';
        commentsListElement.appendChild(li);
        const submitButton = document.getElementById("submit-comment");
        turnOffSubmitButton(submitButton);
        return;
    } */

    if (!comments || !Array.isArray(comments) || comments.length === 0) {
        // If no comments, display a message
        const li = document.createElement('li');
        li.textContent = 'No comments available for this listing.';
        commentsListElement.appendChild(li);
        return;
    }

    // Populate the comments list
    comments.forEach(comment => {
        const li = document.createElement('li');
        // Convert Unix second timestamp to readable date or time
        let dateStr = 'Unknown date';
        // Check if timestamp exists and is a valid int64 in seconds
        if (comment.timestamp !== undefined && comment.timestamp !== null && !isNaN(Number(comment.timestamp))) {
            // Convert int64 microseconds to milliseconds for JS Date
            const dateObj = new Date(comment.timestamp / 1000);

            const now = new Date();

            const diffMs = now - dateObj;

            const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

            const isToday = dateObj.getFullYear() === now.getFullYear() &&
                    dateObj.getMonth() === now.getMonth() &&
                    dateObj.getDate() === now.getDate();

            if (isToday) {
            dateStr = dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
            } else if (diffDays === 1) {
            dateStr = "1 day ago";
            } else if (diffDays > 1 && diffDays < 7) {
            dateStr = `${diffDays} days ago`;
            } else if (diffDays >= 7) {
            dateStr = dateObj.toLocaleDateString();
            }
        }
        li.innerHTML = `<strong>${comment.username}</strong> <span class="comment-date">${dateStr}</span><br>${comment.comment_text}`;
        li.className = 'comment-item';
        commentsListElement.appendChild(li);
    });
}

// Function to show the current URL in the comments tab
function displayURL() {
    const commentsTabContent = document.getElementById('comments');
    if (commentsTabContent) {
        const urlHeader = document.createElement('div');
        urlHeader.className = 'comments-url-header';
        
        // Get the current tab's URL using the Chrome extension API
        chrome.tabs.query({ currentWindow: true, active: true }, function (tabs) {
            console.log(tabs[0].url);
            if (tabs[0] && tabs[0].url) {
                urlHeader.textContent = `Comments for: ${tabs[0].url}`;
            } else {
                urlHeader.textContent = 'Comments for: [unknown URL]';
            }
        });
        // Insert header at the top of the comments tab content
        commentsTabContent.insertBefore(urlHeader, commentsTabContent.firstChild);
    }
}

// Call this function to display the current URL in the comments tab
//displayURL()

// Compiles the form data and user data into a struct for submission
async function handleCommentSubmission(event) {
    // Stop default form submission behavior
    event.preventDefault();

    // Disable the form for 5 seconds to prevent multiple submissions
    const submitButton = event.target.querySelector('button[type="submit"]');
    turnOffSubmitButton(submitButton);
    setTimeout(() => {
        turnOnSubmitButton(submitButton);
    }, 3000);

    // Get comment data
    const username = document.getElementById('username-input').value.trim();
    const commentText = document.getElementById('comment-input').value.trim();
    const listingIdAndType = await getListingIDAndType();
    if (!listingIdAndType) {
        console.error("No valid listing ID or type found in the current URL.");
        return;
    }

    // Compile the comment object
    const commentObj = {
        userId: getLocalUserId(),
        listingID: listingIdAndType.listingID,
        username: username,
        commentText: commentText,
        listingType: listingIdAndType.listingType,
    };

    //console.log('Form submission:', commentObj);

    // Display and post the comment
    //displaySubmittedComment(commentObj)
    postComment(commentObj).then(
        (response) => {
            // Log the result or error
            console.log('Comment posted'/*, response, response.body*/);
            

            if (!response.ok) {
                response.text()
                    .then(responseBody => JSON.parse(responseBody))
                    .then(unmarshalledBody => setErrorMessage(unmarshalledBody.error))
                    .catch(error => setErrorMessage(error));
            } else {
                hideErrorField();
            }

            // Display the updated comments list after posting
            getCommentsByListingId(listingIdAndType, displayComments)
    }).then(result => console.log("Got result "+result)).catch(
        error => {
            console.log("Got error when posting comment to "+`${API_URL}/comments`+": "+error);
            return;
        });

    /* .then(response => response.text())
        .then(result => callbackFunc(result))
        .catch(error => callbackFunc(null, error)); */

    // Clear the form fields after submission
    document.getElementById('comment-input').value = '';
}

// Gets the listing ID from the current URL
// zillow's URL format is "https://www.zillow.com/homedetails/listing-street-name/1234567890_zpid/", 
// from which you would extract "1234567890"
async function getListingIDAndType() {
    let listingURL = '';

    if (CURRENT_TABS[0] != null) {
        listingURL = CURRENT_TABS[0].url;
    } else {
        // Get the current tab's URL using the Chrome extension API
        const tabs = await chrome.tabs.query({ currentWindow: true, active: true });
        //console.log("Current tabs:", tabs);
        if (tabs.length > 0 && tabs[0].url) {
            listingURL = tabs[0].url;
        } else {
            throw new Error("Failed to get current URL");
        }
    }

    // Determine what type of listing it is (house or apartment)

    const houseRegex = RegExp("^https:\/\/www\.zillow\.com\/homedetails\/.*");
    const apartmentRegex = RegExp("^https:\/\/www\.zillow\.com\/apartments\/.*$");

    let listingID;

    if (houseRegex.test(listingURL)) {
        // Extract the listing ID from the URL
        // Find section of the URL that ends with "_zpid"
        const urlParts = listingURL.split('/');
        const zpidIndex = urlParts.findIndex(part => part.endsWith('_zpid'));
        
        if (zpidIndex !== -1 && urlParts[zpidIndex]) {
            // Gets the listing ID by removing the "_zpid" suffix
            listingID = urlParts[zpidIndex].replace('_zpid', '');
            //console.log("Listing ID found:", listingID);
            //console.log({listingID: listingID, listingType: "house"});
            return {listingID: listingID, listingType: "house"};
        }
    } else if (apartmentRegex.test(listingURL)) {
        const urlParts = listingURL.split('/');
        const aptIdIndex = urlParts.length-2;
        
        listingID = urlParts[aptIdIndex];
        if (listingID.length == 6) {
            //console.log({listingID: listingID, listingType: "apt"});
            return {listingID: listingID, listingType: "apt"};
        }
    } 
    
    // If no valid listing ID is found, return null
    // Disable the submit button and show an error message
    const submitButton = document.querySelector('#comment-form button[type="submit"]');
    if (submitButton) {
        turnOffSubmitButton(submitButton);
    }
    console.error("No valid listing ID found in the current URL.");
    return null; // No valid listing ID found
}

// Displays the submitted comment in the extension popup
function displaySubmittedComment(commentObj) {
    const form = document.getElementById('comment-form');
    if (!form) return;

    let submittedSection = document.getElementById('submitted-comment');
    if (!submittedSection) {
        submittedSection = document.createElement('div');
        submittedSection.id = 'submitted-comment';
        form.parentNode.insertBefore(submittedSection, form.nextSibling);
    }

    submittedSection.innerHTML = `
        <h4>Submitted Comment</h4>
        <div><strong>Username:</strong> ${commentObj.username}</div>
        <div><strong>Comment:</strong> ${commentObj.commentText}</div>
        <div><strong>Page:</strong> ${commentObj.listingId}</div>
        <div><strong>User ID:</strong> ${commentObj.userId}</div>
    `;
}

// Fetches the list of comments for a specific listing from the API
function getCommentsByListingId(listingIdAndType, callbackFunc) {
    if (!listingIdAndType) {
        console.error("No valid listing ID or type provided.");
        return [];
    }

    var requestOptions = {
    method: 'GET',
    redirect: 'follow'
    };

    fetch(`${API_URL}/comments/${listingIdAndType.listingID}?listing_type=${listingIdAndType.listingType}`, requestOptions)
        .then(response => response.text())
        .then(result => callbackFunc(result))
        .catch(error => callbackFunc(null, error));
}

// Posts a new comment to the API
async function postComment(commentObj, callbackFunc) {
    // Save the username to localStorage
    saveUsername(commentObj.username);

    // Collect comment data
    let listingTitle = await getListingTitle();

    // Already should be in "commentObj"
    /* let listingIdAndType = await getListingIDAndType();
    if (!listingIdAndType) {
        console.error("No valid listing ID or type found in the current URL.");
        return;
    } */

    // Prepare form data for API
    var myHeaders = new Headers();
    myHeaders.append("Content-Type", "application/x-www-form-urlencoded");

    var urlencoded = new URLSearchParams();
    urlencoded.append("listing_title", listingTitle);
    urlencoded.append("listing_id", commentObj.listingID);
    urlencoded.append("user_id", getLocalUserId());
    urlencoded.append("username", commentObj.username);
    urlencoded.append("comment_text", commentObj.commentText);
    urlencoded.append("listing_type", commentObj.listingType)

    var requestOptions = {
    method: 'POST',
    headers: myHeaders,
    body: urlencoded,
    redirect: 'follow'
    };

    // Send POST request to the API
    return fetch(`${API_URL}/comments`, requestOptions)
}

// getNewUserId retrieves a new V7 (Time-based) UUID from the API
function getNewUserId(callbackFunc) {
    var requestOptions = {
        method: 'GET',
        redirect: 'follow'
    };

    fetch(`${API_URL}/user/user_id`, requestOptions)
        .then(response => response.text())
        .then(result => callbackFunc(result))
        .catch(error => callbackFunc(null, error));
}

// Saves the current username to localStorage
function saveUsername(username) {
    if (username && username.trim() !== '') {
        window.localStorage.setItem('zillow_commenter_username', username.trim());
    } else {
        console.warn('Invalid username provided, not saving to localStorage.');
    }
}

// Retrieves the saved username from localStorage
function getSavedUsername() {
    return window.localStorage.getItem('zillow_commenter_username') || '';
}

// Sets error message
function setErrorMessage(errorText) {
    const errorField = document.getElementById("error-field");
    const errorMessage = document.getElementById("error-message");

    errorMessage.innerHTML = errorText;

    errorField.style.visibility = "visible";
}

// Hide error message 
function hideErrorField() {
    const errorField = document.getElementById("error-field");

    errorField.style.visibility = "hidden";
}


function initializeCommentInputValidation() {
    // Add event listener for comment input validation (vanilla JS, multi-line support)
    document.getElementById('comment-input').addEventListener('keyup', function validateCommentInput() {
        const errorMsg = "Must be letters, numbers, or punction.";
        const textarea = this;
        const pattern = new RegExp(textarea.getAttribute('pattern'));
        let hasError = false;

        // Print helpful debugging data
        /* console.log("Regex pattern:",pattern);
        // Print textarea.value encoded in ASCII hexadecimal
        const asciiHex = Array.from(textarea.value)
            .map(char => char.charCodeAt(0).toString(16).padStart(2, '0'))
            .join(' ');
        console.log("ASCII hex encoding:", asciiHex); */

        // Check the regex pattern against the whole comment text
        if (!pattern.test(textarea.value)) {
            hasError = true;
            console.log("Comment text has an error:\n\n",textarea.value,"\n");
        }

        /* Validate each line separately  (OLD METHOD)
        const lines = textarea.value.split("\n");
        let index = 0;
        for (let line of lines) {
            index++;
            if (line == "") {
                continue;
            }
            if (!pattern.test(line)) {
                hasError = true;
                console.log("Line #",index,":'"+line+"' has an error.");
                break;
            }
        } */
        //console.log("Comment field is valid:",!hasError);
        if (typeof textarea.setCustomValidity === 'function') {
            textarea.setCustomValidity(hasError ? errorMsg : '');
        } else {
            textarea.classList.toggle('error', hasError);
            textarea.classList.toggle('ok', !hasError);
            if (hasError) {
                textarea.title = errorMsg;
            } else {
                textarea.removeAttribute('title');
            }
        }
    });
}

function initializeZillowetteIcon() {
    // Modal popup logic for zillowette icon
    const icon = document.getElementById('zillowette-icon');
    const modal = document.getElementById('icon-modal');
    const input = document.getElementById('icon-modal-input');
    const submitBtn = document.getElementById('icon-modal-submit');
    const cancelBtn = document.getElementById('icon-modal-cancel');

    if (!icon || !modal || !input || !submitBtn || !cancelBtn) return;

    modal.style.display = 'none';

    icon.addEventListener('click', function() {
        if (modal.style.display != 'none') {
            modal.style.display = 'none';
        } else {
            modal.style.display = 'flex';
        }
        input.value = '';
        input.focus();
    });

    cancelBtn.addEventListener('click', function() {
        modal.style.display = 'none';
    });

    submitBtn.addEventListener('click', function() {
        const value = input.value.trim();
        if (value !== '') {
            // Set API address to input value
            console.log("API URL set to:",setApiAddress(value));
        }
        modal.style.display = 'none';
    });

    // Optional: submit on Enter key
    input.addEventListener('keydown', function(e) {
        if (e.key === 'Enter') {
            submitBtn.click();
        }
    });
}

function setApiAddress(apiAddress) {
    API_ADDRESS = apiAddress;
    API_URL = `${API_ADDRESS}/api/v1`;
    
    // Saves API address to local storage
    window.localStorage.setItem("zillow_commenter_api_address", API_ADDRESS);

    // Setting API address automatically refreshes comments
    populateComments();
    return API_URL;
}

// Gets listing title from the injected content script
async function getListingTitle() {
    /* const {statusCode, title} = await chrome.runtime.sendMessage({
        action: 'get_listing_title'
    }); */

    // Get current tab ID
    let currentTabID;

    if (CURRENT_TABS[0] != null) {
        currentTabID = CURRENT_TABS[0].id;
    } else {
        const currentTabs = await chrome.tabs.query({active: true, lastFocusedWindow: true});  

        if (currentTabs[0] == null) { 
            console.log("Current tabs:",currentTabs)
            console.error("Current tab is null when querying for listing title.");
            throw new Error("Current tab is null when querying for listing title.");
        } 

        currentTabID = currentTabs[0].id
    }
    

    // Query current tab's content script for a title
    //console.log("Got current tabs:",tabs)
    const response = await chrome.tabs.sendMessage(currentTabID, {action: "get_listing_title"});

    if (chrome.runtime.lastError) {
        // Called when an error occurs in getting the title
        console.error("Content script not available:", chrome.runtime.lastError.message);
        throw new Error("Content script not available:"+chrome.runtime.lastError.message);
    } 

    if (!response.title) {
        console.error("Title not available in call.");
        throw new Error("Title not available in call.");
    }

    // Called when response is recieved from content script
    console.log("Got title:",response.title);

    if (response.title === undefined) {
        console.error("Title is undefined");
        throw new Error("Title is undefined");
    }

    return sanitizeListingTitle(response.title);
}

// Sanitizes the listing title to only include printable ascii characters
function sanitizeListingTitle(listingTitle) {
    if (typeof listingTitle !== 'string') return '';
    // Printable ASCII: 32 (space) to 126 (~)
    return listingTitle.replace(/[^\x20-\x7E]+/g, '');
}

function turnOffSubmitButton(submitButton) {
    submitButton.disabled = true;
    submitButton.style.backgroundColor = '#ccc';
}

function turnOnSubmitButton(submitButton) {
    submitButton.disabled = false;
    submitButton.style.backgroundColor = ''; 
}