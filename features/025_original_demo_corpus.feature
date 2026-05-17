Feature: Original demo corpus

Background:
  Given the original demo corpus task is accepted

Scenario Outline: original demos are imported from a documented source
  When the coder imports the original XSpringies demo corpus
  Then imported demo file <demo_file> should exist under <demo_directory>
  And imported demo file <demo_file> should preserve its original filename

Examples:
  | demo_directory | demo_file          |
  | demos/original | 2snake.xsp         |
  | demos/original | 2x2snake.xsp       |
  | demos/original | 3snake.xsp         |
  | demos/original | 4snake.xsp         |
  | demos/original | 9snake.xsp         |
  | demos/original | ball.xsp           |
  | demos/original | belt-loop.xsp      |
  | demos/original | belt-tire.xsp      |
  | demos/original | belt.xsp           |
  | demos/original | big-jello.xsp      |
  | demos/original | bike.xsp           |
  | demos/original | billiard.xsp       |
  | demos/original | blanket.xsp        |
  | demos/original | bowl.xsp           |
  | demos/original | bridge.xsp         |
  | demos/original | diamond-chain.xsp  |
  | demos/original | hammer.xsp         |
  | demos/original | hexball1.xsp       |
  | demos/original | hexball2.xsp       |
  | demos/original | hexball3.xsp       |
  | demos/original | hexhexmesh.xsp     |
  | demos/original | hexmesh.xsp        |
  | demos/original | jello-buttress.xsp |
  | demos/original | jello-pend.xsp     |
  | demos/original | jello.xsp          |
  | demos/original | jello2.xsp         |
  | demos/original | kalied-maker.xsp   |
  | demos/original | kalied1.xsp        |
  | demos/original | kalied2.xsp        |
  | demos/original | lgrid.xsp          |
  | demos/original | lissajous.xsp      |
  | demos/original | loopy.xsp          |
  | demos/original | mesh.xsp           |
  | demos/original | nifty.xsp          |
  | demos/original | nifty12.xsp        |
  | demos/original | octanifty.xsp      |
  | demos/original | orbit1.5.xsp       |
  | demos/original | orbit2.xsp         |
  | demos/original | pend.xsp           |
  | demos/original | pend7x1.xsp        |
  | demos/original | pend7x2.xsp        |
  | demos/original | pendbees.xsp       |
  | demos/original | pendwave.xsp       |
  | demos/original | person1.xsp        |
  | demos/original | person2.xsp        |
  | demos/original | plane.xsp          |
  | demos/original | psycho.xsp         |
  | demos/original | serp.xsp           |
  | demos/original | slide.xsp          |
  | demos/original | snake.xsp          |
  | demos/original | square.xsp         |
  | demos/original | stretchmesh.xsp    |
  | demos/original | super-jello.xsp    |
  | demos/original | tire.xsp           |
  | demos/original | transpend.xsp      |
  | demos/original | tri4.xsp           |
  | demos/original | trimesh.xsp        |
  | demos/original | urchin.xsp         |
  | demos/original | wave.xsp           |
  | demos/original | worm.xsp           |
  | demos/original | xgrid.xsp          |
  | demos/original | xtrimesh.xsp       |
  | demos/original | zgrid.xsp          |
  | demos/original | zharp.xsp          |
  | demos/original | zharp2.xsp         |
  | demos/original | zingy.xsp          |
  | demos/original | zwave.xsp          |

Scenario Outline: every imported original demo loads
  Given imported original demo file <demo_file> exists
  When the coder loads imported original demo file <demo_file>
  Then loading should pass

Examples:
  | demo_file          |
  | 2snake.xsp         |
  | 2x2snake.xsp       |
  | 3snake.xsp         |
  | 4snake.xsp         |
  | 9snake.xsp         |
  | ball.xsp           |
  | belt-loop.xsp      |
  | belt-tire.xsp      |
  | belt.xsp           |
  | big-jello.xsp      |
  | bike.xsp           |
  | billiard.xsp       |
  | blanket.xsp        |
  | bowl.xsp           |
  | bridge.xsp         |
  | diamond-chain.xsp  |
  | hammer.xsp         |
  | hexball1.xsp       |
  | hexball2.xsp       |
  | hexball3.xsp       |
  | hexhexmesh.xsp     |
  | hexmesh.xsp        |
  | jello-buttress.xsp |
  | jello-pend.xsp     |
  | jello.xsp          |
  | jello2.xsp         |
  | kalied-maker.xsp   |
  | kalied1.xsp        |
  | kalied2.xsp        |
  | lgrid.xsp          |
  | lissajous.xsp      |
  | loopy.xsp          |
  | mesh.xsp           |
  | nifty.xsp          |
  | nifty12.xsp        |
  | octanifty.xsp      |
  | orbit1.5.xsp       |
  | orbit2.xsp         |
  | pend.xsp           |
  | pend7x1.xsp        |
  | pend7x2.xsp        |
  | pendbees.xsp       |
  | pendwave.xsp       |
  | person1.xsp        |
  | person2.xsp        |
  | plane.xsp          |
  | psycho.xsp         |
  | serp.xsp           |
  | slide.xsp          |
  | snake.xsp          |
  | square.xsp         |
  | stretchmesh.xsp    |
  | super-jello.xsp    |
  | tire.xsp           |
  | transpend.xsp      |
  | tri4.xsp           |
  | trimesh.xsp        |
  | urchin.xsp         |
  | wave.xsp           |
  | worm.xsp           |
  | xgrid.xsp          |
  | xtrimesh.xsp       |
  | zgrid.xsp          |
  | zharp.xsp          |
  | zharp2.xsp         |
  | zingy.xsp          |
  | zwave.xsp          |

Scenario Outline: imported original demos remain separate from starter demos
  Given starter demo file <starter_demo> exists
  When the coder imports the original XSpringies demo corpus
  Then starter demo file <starter_demo> should remain under <starter_directory>
  And original demos should remain under <original_directory>

Examples:
  | starter_directory | starter_demo     | original_directory |
  | demos             | pendulum.xsp     | demos/original     |
  | demos             | spring-chain.xsp | demos/original     |
  | demos             | small-mesh.xsp   | demos/original     |

Scenario Outline: demo provenance is documented
  When the coder imports the original XSpringies demo corpus
  Then provenance field <field> should be documented

Examples:
  | field           |
  | source URL      |
  | retrieval date  |
  | package version |
  | license context |
