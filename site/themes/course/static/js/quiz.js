/* Interactive lesson quiz: renders quiz.json embedded by the quiz shortcode.
   Vanilla JS, no dependencies. Safe to load more than once. */
(function () {
  "use strict";
  if (window.__coursesmithQuizLoaded) return;
  window.__coursesmithQuizLoaded = true;

  function el(tag, className, text) {
    var node = document.createElement(tag);
    if (className) node.className = className;
    if (text) node.textContent = text;
    return node;
  }

  function renderQuiz(container) {
    var dataNode = container.querySelector("[data-quiz-data]");
    if (!dataNode) return;
    var quiz;
    try {
      quiz = JSON.parse(dataNode.textContent);
    } catch (err) {
      container.appendChild(el("p", "quiz-error", "This quiz could not be loaded."));
      return;
    }

    var form = el("form", "quiz-form");
    form.setAttribute("novalidate", "");
    form.appendChild(el("h2", "quiz-title", quiz.title || "Check your understanding"));

    quiz.questions.forEach(function (q, qi) {
      var fieldset = el("fieldset", "quiz-question");
      fieldset.dataset.answer = String(q.answer_index);
      fieldset.appendChild(el("legend", "quiz-prompt", (qi + 1) + ". " + q.prompt));

      q.options.forEach(function (option, oi) {
        var label = el("label", "quiz-option");
        var input = document.createElement("input");
        input.type = "radio";
        input.name = "q-" + qi;
        input.value = String(oi);
        label.appendChild(input);
        label.appendChild(el("span", null, option));
        fieldset.appendChild(label);
      });

      var explanation = el("p", "quiz-explanation", q.explanation || "");
      explanation.hidden = true;
      fieldset.appendChild(explanation);
      form.appendChild(fieldset);
    });

    var submit = el("button", "quiz-submit", "Check answers");
    submit.type = "submit";
    form.appendChild(submit);
    var result = el("p", "quiz-result");
    result.setAttribute("role", "status");
    form.appendChild(result);

    form.addEventListener("submit", function (event) {
      event.preventDefault();
      var questions = form.querySelectorAll(".quiz-question");
      var correct = 0;
      var answered = 0;

      questions.forEach(function (fieldset) {
        var chosen = fieldset.querySelector("input:checked");
        fieldset.classList.remove("is-correct", "is-incorrect");
        var explanation = fieldset.querySelector(".quiz-explanation");
        explanation.hidden = true;
        if (!chosen) return;
        answered++;
        if (chosen.value === fieldset.dataset.answer) {
          correct++;
          fieldset.classList.add("is-correct");
        } else {
          fieldset.classList.add("is-incorrect");
        }
        explanation.hidden = false;
      });

      if (answered < questions.length) {
        result.textContent = "Answer all " + questions.length + " questions first (" +
          (questions.length - answered) + " left).";
        return;
      }
      result.textContent = "You got " + correct + " of " + questions.length + " right." +
        (correct === questions.length ? " 🎉" : " Review the explanations above and try again.");
    });

    container.appendChild(form);
    container.dataset.quizReady = "true";
  }

  function init() {
    document.querySelectorAll("[data-quiz]:not([data-quiz-ready])").forEach(renderQuiz);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
