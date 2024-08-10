async function onClickSavedJobs(event, htmlElem) {
  event.preventDefault();
  if (!htmlElem) {
    htmlElem = this;
  }
  if (!(htmlElem instanceof HTMLElement)) {
    throw new Error("htmlElem onClickSavedJobs is not HTMLElement ");
  }
  const svgElem = htmlElem.querySelector("svg.favorite-icon");
  if (!(svgElem instanceof SVGElement)) {
    throw new Error("svgElem onClickSavedJobs is not HTMLElement ");
  }
  if (!svgElem.classList.contains("login")) {
    svgElem.classList.toggle("saved");
    svgElem.classList.toggle("unsaved");
  }
  let hxVals = htmlElem.getAttribute("hx-vals");
  let path = htmlElem.getAttribute("path");
  if (!hxVals || hxVals.length <= 0) {
    window.location.replace(path);
    return;
  }
  const params = new URLSearchParams();
  if (hxVals.length > 0) {
    let tempHxvals = JSON.parse(hxVals);
    for (let [key, value] of Object.entries(tempHxvals)) {
      params.append(key, value);
    }
  }

  const respone = await fetch(path, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body: params,
  });
  const resText = await respone.text();
  if (!respone.ok) {
    svgElem.classList.toggle("saved");
    svgElem.classList.toggle("unsaved");
    console.error(resText);
    return;
  }
  if (htmlElem && document.body.contains(htmlElem)) {
    htmlElem.insertAdjacentHTML("beforebegin", resText);
    htmlElem.remove();
  }
}
function fetchJobsRefresher() {
  $("form.formFetchJobs").waypoint({
    offset: "100%",
    handler: async function () {
      try {
        const formElem = this.element;
        if (!(formElem instanceof HTMLFormElement)) {
          throw new Error("formElem is not a HTMLFormELement");
        }
        this.destroy();
        fetchJobs(formElem);
      } catch (error) {
        console.error(error);
      }
    },
  });
}

$(document).ready(function () {
  $("#searchJobsInp").val("");
  $("#searchJobsInp").on("input", function () {
    const input = $("#searchJobsInp").val();
    const arr = ["computrabajo", "linkedin", "indeed"];
    let str = `{ "data": { `;

    arr.forEach((job) => {
      str += `"${job}": {"query": "${input}", "page": "1"},`;
    });

    str = str.slice(0, -1);
    str += `}, "sites": "${arr.join(",")}" }`;

    $("#fetchJobsForm").attr("hx-vals", str);
  });
  $("#fetchJobsForm").submit(function (e) {
    e.preventDefault();
    Waypoint.destroyAll();
    $("#container").html("");
  });
});

async function _fetchJobs(formElem) {
  return new Promise(async (resolve, reject) => {
    if (!formElem) {
      formElem = this;
    }
    let hxIndicator;
    let hxExt;
    try {
      if (!(formElem instanceof HTMLElement)) {
        throw new Error("htmlELem is not a HTMLELement");
      }
      let hxVals = formElem.getAttribute("hx-vals");
      hxExt = formElem.getAttribute("hx-ext");
      let targetSelector = formElem.getAttribute("hx-target");
      let hxIndicatorSelector = formElem.getAttribute("hx-indicator");
      let target = document.querySelector(targetSelector);
      hxIndicator = document.querySelector(hxIndicatorSelector);
      if (!hxVals || hxVals.length <= 0) {
        throw new Error("hxVals has no value");
      }
      if (hxExt === "remove") formElem.remove();
      if (hxIndicator) {
        hxIndicator.classList.add("htmx-request");
      }

      const response = await fetch(`/jobs/get`, {
        method: "POST",
        body: hxVals,
        headers: {
          "Content-Type": "application/json",
        },
      });
      if (!response.ok) {
        console.error(await response.text());
        return resolve();
      }
      const reader = response.body.getReader();
      const decoder = new TextDecoder("utf-8");
      let chunk;
      let savedData = "";

      while (!(chunk = await reader.read()).done) {
        let data = decoder.decode(chunk.value, { stream: true });
        if (matchFirst(data.trim(), "<a") && !data.trim().includes("</a>")) {
          savedData = data;
          continue;
        }
        if (
          !matchFirst(data.trim(), "<a") &&
          matchLast(data.trim(), "</a>") &&
          !savedData.length > 0
        ) {
          continue;
        } else if (
          !matchFirst(data.trim(), "<a") &&
          matchLast(data.trim(), "</a>")
        ) {
          data = savedData + data;
          savedData = "";
        }
        if (target instanceof HTMLElement) {
          target.insertAdjacentHTML("beforeend", data);
        }
        if (hxIndicator instanceof HTMLElement) {
          hxIndicator.classList.remove("htmx-request");
        }
      }
      fetchJobsRefresher();
      resolve();
    } catch (error) {
      console.error(error);
      reject(error);
    }
    if (hxExt === "remove") formElem.remove();
    if (hxIndicator instanceof HTMLElement) {
      hxIndicator.classList.remove("htmx-request");
    }
    resolve();
  });
}

function matchEdges(str, trgFirst, trgSecond) {
  return matchFirst(str, trgFirst) && matchLast(str, trgSecond);
}

function matchLast(str, target) {
  const lenTarget = target.length;
  const lenStr = str.length;
  if (lenTarget > lenStr) {
    return false;
  }
  return str.substring(lenStr - lenTarget) === target;
}

function matchFirst(str, target) {
  const lenTarget = target.length;
  const lenStr = str.length;
  if (lenTarget > lenStr) {
    return false;
  }
  return str.substring(0, lenTarget) === target;
}

function readMore(event, htmlElem) {
  event.preventDefault();
  if (!htmlElem) {
    htmlElem = this;
  }
  try {
    if (!(htmlElem instanceof HTMLElement)) {
      throw new Error("htmlElem is not an HTMLElement");
    }
    const [parentElem, ok] = findClosestParent(
      htmlElem.parentElement,
      "p.text-base"
    );
    if (!ok) {
      throw new Error("parent not found");
    }
    const content = parentElem.getAttribute("content");
    const clicked = parentElem.getAttribute("clicked");
    if (clicked === "true") {
      parentElem.setAttribute("clicked", "false");
      htmlElem.textContent = "read more...";
      parentElem.textContent = content.slice(0, 200) + "...";
      parentElem.insertAdjacentElement("beforeend", htmlElem);
      return;
    }
    if (clicked === "false") {
      parentElem.setAttribute("clicked", "true");
      htmlElem.textContent = "read less";
      parentElem.textContent = content;
      parentElem.insertAdjacentElement("beforeend", htmlElem);
      return;
    }
  } catch (error) {
    console.error(error);
  }
}

function findClosestParent(htmlElem, target) {
  if (!(htmlElem instanceof HTMLElement)) {
    throw new Error("htmlElem is not an HTMLElement");
  }
  if (typeof target !== "string" || target.length <= 0) {
    throw new Error("target is not a string or is an empty string");
  }
  if (htmlElem == document.body) return [null, false];
  const [tagName, ...classes] = target.split(".");

  if (
    classes.length === 0 &&
    tagName.toLowerCase() === htmlElem.tagName.toLowerCase()
  ) {
    return [htmlElem, true];
  }
  if (
    classes.every((cls) => htmlElem.classList.contains(cls)) &&
    tagName.toLowerCase() === htmlElem.tagName.toLowerCase()
  ) {
    return [htmlElem, true];
  }

  return findClosestParent(htmlElem.parentElement, target);
}

function debounceRequest(fn) {
  let reqActive = false;

  return function (...args) {
    if (reqActive) return;

    reqActive = true;
    fn(...args).finally(() => {
      reqActive = false;
    });
  };
}

function debounce(fn, immediate = false, delay = 200) {
  let timeout;
  let ok = false;
  return function (...args) {
    if (immediate && !ok) {
      fn(...args);
      ok = true;
      return;
    }
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      if (!immediate || ok) {
        fn(...args);
        ok = false;
      }
    }, delay);
  };
}

const fetchJobs = debounceRequest(_fetchJobs);