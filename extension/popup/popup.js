console.log("User ID: ", getUserId());

// Sets a unique user ID in localStorage if it doesn't exist
//
// Note: The localstorage persists between browser sessions, and incognito mode
function setUserId() {
    let userId = window.localStorage.getItem('zillow_commenter_user_id');
    if (!userId) {
        userId = "user_" + Math.random().toString(36).substring(2, 15);
        window.localStorage.setItem('zillow_commenter_user_id', userId);
    }
}

//setUserId();

// Retrieves the user ID from localStorage
function getUserId() {
    return window.localStorage.getItem('zillow_commenter_user_id');
}

//console.log(getUserId());


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

// Sample comments array (just strings)
const sampleComments = [
    "Great property, loved the backyard!",
    "Needs some renovation, but has potential.",
    "Amazing location and spacious rooms.",
    "Too expensive for the area.",
    "Would love to schedule a tour."
];

// Sample comment objects
const sampleCommentObjects = [
    {
        datePosted: "2024-06-01",
        username: "user123",
        commentText: "Great property, loved the backyard!"
    },
    {
        datePosted: "2024-06-02",
        username: "houseHunter",
        commentText: "Needs some renovation, but has potential."
    },
    {
        datePosted: "2024-06-03",
        username: "realtyFan",
        commentText: "Amazing location and spacious rooms."
    },
    {
        datePosted: "2024-06-04",
        username: "skeptic42",
        commentText: "Too expensive for the area."
    },
    {
        datePosted: "2024-06-05",
        username: "tourSeeker",
        commentText: "Would love to schedule a tour."
    }
];

// Function to populate comments list in the DOM
function populateComments(comments) {
    const commentsList = document.querySelector('.comments-list');
    console.log('Populating comments:', comments, commentsList);
    if (!commentsList) return;
    commentsList.innerHTML = '';
    comments.forEach(comment => {
        const li = document.createElement('li');
        li.innerHTML = `<strong>${comment.username}</strong> (${comment.datePosted}):<br>${comment.commentText}`;
        commentsList.appendChild(li);
    });
}

// Example usage: populate with sample comments
populateComments(sampleCommentObjects);

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

// Call the function to display the current URL in the comments tab
//displayURL()


document.getElementById('comment-form').addEventListener('submit', handleCommentSubmit);

// Compiles the form data and user data into a struct for submission
async function handleCommentSubmit(event) {
    event.preventDefault();
    const username = document.getElementById('username-input').value.trim();
    const commentText = document.getElementById('comment-input').value.trim();
    const datePosted = new Date().toISOString().split('T')[0];

    let targetZillowPage = '';
    // Get the current tab's URL using the Chrome extension API
    const tabs = await chrome.tabs.query({ currentWindow: true, active: true });
    if (tabs.length > 0 && tabs[0].url) {
        targetZillowPage = tabs[0].url;
    }

    const commentObj = {
        userId: getUserId(),
        target_zillow_page: targetZillowPage,
        datePosted,
        username,
        commentText
    };
    //console.log('Form submission:', commentObj);
    displaySubmittedComment(commentObj)
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
        <div><strong>Date:</strong> ${commentObj.datePosted}</div>
        <div><strong>Comment:</strong> ${commentObj.commentText}</div>
        <div><strong>Page:</strong> ${commentObj.target_zillow_page}</div>
        <div><strong>User ID:</strong> ${commentObj.userId}</div>
    `;

    sampleCommentObjects.push(commentObj);
    populateComments(sampleCommentObjects);
}