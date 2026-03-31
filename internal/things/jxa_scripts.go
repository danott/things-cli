package things

// JXA scripts that return JSON. Each script is a template with %s placeholders
// for parameters that get escaped via escapeJS before interpolation.

const listTodosScript = `
const things = Application("Things3");
const todos = things.lists.byName("%s").toDos();
JSON.stringify(todos.map(t => {
  const proj = t.project();
  const area = t.area();
  return {
    id: t.id(),
    name: t.name(),
    status: t.status(),
    notes: t.notes() || "",
    tagNames: t.tagNames() || "",
    dueDate: t.dueDate() ? t.dueDate().toISOString() : null,
    activationDate: t.activationDate() ? t.activationDate().toISOString() : null,
    creationDate: t.creationDate() ? t.creationDate().toISOString() : null,
    modificationDate: t.modificationDate() ? t.modificationDate().toISOString() : null,
    completionDate: t.completionDate() ? t.completionDate().toISOString() : null,
    cancellationDate: t.cancellationDate() ? t.cancellationDate().toISOString() : null,
    projectName: proj ? proj.name() : "",
    areaName: area ? area.name() : ""
  };
}));
`

const projectTodosScript = `
const things = Application("Things3");
const project = things.projects.byName("%s");
const todos = project.toDos();
JSON.stringify(todos.map(t => ({
  id: t.id(),
  name: t.name(),
  status: t.status(),
  notes: t.notes() || "",
  tagNames: t.tagNames() || "",
  dueDate: t.dueDate() ? t.dueDate().toISOString() : null,
  activationDate: t.activationDate() ? t.activationDate().toISOString() : null,
  creationDate: t.creationDate() ? t.creationDate().toISOString() : null,
  modificationDate: t.modificationDate() ? t.modificationDate().toISOString() : null,
  completionDate: t.completionDate() ? t.completionDate().toISOString() : null,
  cancellationDate: t.cancellationDate() ? t.cancellationDate().toISOString() : null,
  projectName: "%s",
  areaName: ""
})));
`

const areaTodosScript = `
const things = Application("Things3");
const area = things.areas.byName("%s");
const todos = area.toDos();
JSON.stringify(todos.map(t => {
  const proj = t.project();
  return {
    id: t.id(),
    name: t.name(),
    status: t.status(),
    notes: t.notes() || "",
    tagNames: t.tagNames() || "",
    dueDate: t.dueDate() ? t.dueDate().toISOString() : null,
    activationDate: t.activationDate() ? t.activationDate().toISOString() : null,
    creationDate: t.creationDate() ? t.creationDate().toISOString() : null,
    modificationDate: t.modificationDate() ? t.modificationDate().toISOString() : null,
    completionDate: t.completionDate() ? t.completionDate().toISOString() : null,
    cancellationDate: t.cancellationDate() ? t.cancellationDate().toISOString() : null,
    projectName: proj ? proj.name() : "",
    areaName: "%s"
  };
}));
`

const tagTodosScript = `
const things = Application("Things3");
const tag = things.tags.byName("%s");
const todos = tag.toDos();
JSON.stringify(todos.map(t => {
  const proj = t.project();
  const area = t.area();
  return {
    id: t.id(),
    name: t.name(),
    status: t.status(),
    notes: t.notes() || "",
    tagNames: t.tagNames() || "",
    dueDate: t.dueDate() ? t.dueDate().toISOString() : null,
    activationDate: t.activationDate() ? t.activationDate().toISOString() : null,
    creationDate: t.creationDate() ? t.creationDate().toISOString() : null,
    modificationDate: t.modificationDate() ? t.modificationDate().toISOString() : null,
    completionDate: t.completionDate() ? t.completionDate().toISOString() : null,
    cancellationDate: t.cancellationDate() ? t.cancellationDate().toISOString() : null,
    projectName: proj ? proj.name() : "",
    areaName: area ? area.name() : ""
  };
}));
`

const getTodoScript = `
const things = Application("Things3");
const t = things.toDos.byId("%s");
const proj = t.project();
const area = t.area();
JSON.stringify({
  id: t.id(),
  name: t.name(),
  status: t.status(),
  notes: t.notes() || "",
  tagNames: t.tagNames() || "",
  dueDate: t.dueDate() ? t.dueDate().toISOString() : null,
  activationDate: t.activationDate() ? t.activationDate().toISOString() : null,
  creationDate: t.creationDate() ? t.creationDate().toISOString() : null,
  modificationDate: t.modificationDate() ? t.modificationDate().toISOString() : null,
  completionDate: t.completionDate() ? t.completionDate().toISOString() : null,
  cancellationDate: t.cancellationDate() ? t.cancellationDate().toISOString() : null,
  projectName: proj ? proj.name() : "",
  areaName: area ? area.name() : ""
});
`

const listProjectsScript = `
const things = Application("Things3");
const projects = things.projects();
JSON.stringify(projects.map(p => ({
  id: p.id(),
  name: p.name(),
  status: p.status(),
  notes: p.notes() || ""
})));
`

const getProjectScript = `
const things = Application("Things3");
const p = things.projects.byId("%s");
JSON.stringify({
  id: p.id(),
  name: p.name(),
  status: p.status(),
  notes: p.notes() || ""
});
`

const listTagsScript = `
const things = Application("Things3");
const tags = things.tags();
JSON.stringify(tags.map(t => ({
  id: t.id(),
  name: t.name(),
  parentTag: t.parentTag() ? t.parentTag().name() : ""
})));
`
