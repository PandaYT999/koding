class NMountItemView extends NFileItemView

  constructor:(options = {},data)->

    options.cssClass  or= "mount"
    super options, data
